package upload_test

import (
	"context"
	"errors"
	"go-shop-backend/config"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/internal/upload"
	uploadMocks "go-shop-backend/internal/upload/mocks"
	"go-shop-backend/pkg/apperror"
	contenttypeMocks "go-shop-backend/pkg/contenttype/mocks"
	"go-shop-backend/pkg/storage"
	storageMocks "go-shop-backend/pkg/storage/mocks"
	"go-shop-backend/pkg/testutils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ManagerTestSuite struct {
	suite.Suite
	storage       *storageMocks.MockStorage
	uploadRepo    *repoMocks.MockUploadRepository
	uploadConfig  config.Upload
	ctDetector    *contenttypeMocks.MockDetector
	handler       *uploadMocks.MockHandler
	registry      *uploadMocks.MockRegistry
	uploadManager upload.Manager

	ctx      context.Context
	uploadID uuid.UUID
	entityID uuid.UUID
	purpose  string
}

func (suite *ManagerTestSuite) SetupTest() {
	suite.storage = storageMocks.NewMockStorage(suite.T())
	suite.uploadRepo = repoMocks.NewMockUploadRepository(suite.T())
	suite.uploadConfig = config.Upload{
		MaxFileSize:     1024 * 1024 * 5,
		PresignedURLTTL: time.Minute,
	}
	suite.ctDetector = contenttypeMocks.NewMockDetector(suite.T())
	suite.handler = uploadMocks.NewMockHandler(suite.T())
	suite.registry = uploadMocks.NewMockRegistry(suite.T())
	suite.uploadManager = upload.New(
		suite.storage,
		suite.uploadRepo,
		suite.uploadConfig,
		suite.ctDetector,
		suite.registry,
		testutils.NewSlogLogger(),
	)

	suite.ctx = context.Background()
	suite.uploadID = uuid.New()
	suite.entityID = uuid.New()
	suite.purpose = "test"
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

// ==================== SignURL Tests ====================

func (suite *ManagerTestSuite) TestSignURL_Success() {
	req := dto.GeneratePresignedURLRequest{
		ContentType: "image/png",
		Entity:      dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		Ext:         "png",
		Purpose:     suite.purpose,
	}

	presignedPost := &storage.TemporaryUploadURL{
		URL: "https://s3.example.com/media",
		Fields: map[string]string{
			"Entity-Type": req.Entity.Type,
			"Entity-Id":   req.Entity.ID.String(),
			"Purpose":     req.Purpose,
		},
	}

	filePolicy := upload.Policy{
		MinSize: 5 << 10,
		MaxSize: 1 << 20,
		AllowedFormats: []upload.Format{
			{
				Extensions: []string{"png"}, ContentType: "image/png",
			},
		},
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{
			Policy:  filePolicy,
			Handler: suite.handler,
		}, nil).Once()

	suite.handler.EXPECT().CanSignURL(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	suite.storage.EXPECT().TemporaryUploadURL(suite.ctx, mock.MatchedBy(func(opts storage.TemporaryUploadURLOptions) bool {
		return opts.ContentType == "image/png" &&
			opts.Metadata["Entity-Id"] == suite.entityID.String() &&
			opts.Metadata["Entity-Type"] == string(models.EntityTypeProduct) &&
			opts.Metadata["Purpose"] == req.Purpose &&
			opts.MinSize == 5<<10 &&
			opts.MaxSize == 1<<20 &&
			!opts.Expires.IsZero()
	})).Return(presignedPost, nil).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(presignedPost.URL, response.UploadURL)
	suite.Equal("image/png", response.ContentType)
}

func (suite *ManagerTestSuite) TestSignURL_Forbidden() {
	req := dto.GeneratePresignedURLRequest{
		ContentType: "image/png",
		Entity:      dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		Ext:         "png",
		Purpose:     suite.purpose,
	}

	filePolicy := upload.Policy{
		MinSize: 5 << 10,
		MaxSize: 1 << 20,
		AllowedFormats: []upload.Format{
			{
				Extensions: []string{"png"}, ContentType: "image/png",
			},
		},
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{
			Policy:  filePolicy,
			Handler: suite.handler,
		}, nil).Once()

	suite.handler.EXPECT().CanSignURL(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(apperror.ErrForbidden).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
	suite.ErrorContains(err, "uploadManager.SignURL")
}

func (suite *ManagerTestSuite) TestSignURL_InvalidPurpose() {
	req := dto.GeneratePresignedURLRequest{
		Purpose: suite.purpose,
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{}, upload.ErrPurposeNotFound).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidPurpose)
	suite.ErrorContains(err, "uploadManager.SignURL")
}

func (suite *ManagerTestSuite) TestSignURL_ContentTypeMismatch() {
	req := dto.GeneratePresignedURLRequest{
		ContentType: "video/mp4",
		Ext:         "mp4",
		Purpose:     suite.purpose,
	}

	filePolicy := upload.Policy{
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
			{Extensions: []string{"jpg", "jpeg"}, ContentType: "image/jpeg"},
		},
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{
			Policy:  filePolicy,
			Handler: suite.handler,
		}, nil).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrContentTypeMismatch)
	suite.ErrorContains(err, "uploadManager.SignURL")
}

func (suite *ManagerTestSuite) TestSignURL_StorageError() {
	req := dto.GeneratePresignedURLRequest{
		Ext:         "png",
		ContentType: "image/png",
		Entity:      dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		Purpose:     suite.purpose,
	}

	filePolicy := upload.Policy{
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
		},
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{
			Policy:  filePolicy,
			Handler: suite.handler,
		}, nil).Once()

	suite.handler.EXPECT().CanSignURL(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	expectedErr := errors.New("storage error")
	suite.storage.EXPECT().TemporaryUploadURL(suite.ctx, mock.Anything).
		Return(nil, expectedErr).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, expectedErr)
	suite.ErrorContains(err, "uploadManager.SignURL")
}

// ==================== Attach Tests ====================

func (suite *ManagerTestSuite) TestAttach_Success() {
	req := dto.AttachFileRequest{
		UploadID:  suite.uploadID,
		ObjectKey: "products/123/img.png",
		Entity:    dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		Purpose:   suite.purpose,
	}

	filePolicy := upload.Policy{
		MaxSize: 5 << 20,
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
		},
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(models.EntityTypeProduct),
		"Purpose":     req.Purpose,
	}

	objectInfo := &storage.ObjectInfo{
		Size:     1024,
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{
			Policy:  filePolicy,
			Handler: suite.handler,
		}, nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objectInfo, nil).Once()

	suite.uploadRepo.EXPECT().ExistsByObjectKey(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.EXPECT().Get(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	detectedCT := "image/png"
	suite.ctDetector.EXPECT().Detect(obj).
		Return(detectedCT, nil).Once()

	suite.handler.EXPECT().CanAttach(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	suite.uploadRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(u *models.Upload) bool {
		return u.ObjectKey == req.ObjectKey &&
			u.EntityID == suite.entityID &&
			u.EntityType == models.EntityTypeProduct &&
			*u.ContentType == detectedCT &&
			u.MediaType != ""
	})).Return(nil).Once()

	url := "https://s3.example.com/img.png"
	suite.storage.EXPECT().PublicURL(req.ObjectKey).
		Return(url).Once()

	response, err := suite.uploadManager.Attach(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(url, response.URL)
}

func (suite *ManagerTestSuite) TestAttach_Forbidden() {
	req := dto.AttachFileRequest{
		UploadID:  suite.uploadID,
		ObjectKey: "products/123/img.png",
		Entity:    dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		Purpose:   suite.purpose,
	}

	filePolicy := upload.Policy{
		MaxSize: 5 << 20,
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
		},
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(models.EntityTypeProduct),
		"Purpose":     req.Purpose,
	}

	objectInfo := &storage.ObjectInfo{
		Size:     1024,
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{
			Policy:  filePolicy,
			Handler: suite.handler,
		}, nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objectInfo, nil).Once()

	suite.uploadRepo.EXPECT().ExistsByObjectKey(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.EXPECT().Get(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	detectedCT := "image/png"
	suite.ctDetector.EXPECT().Detect(obj).
		Return(detectedCT, nil).Once()

	suite.handler.EXPECT().CanAttach(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(apperror.ErrForbidden).Once()

	response, err := suite.uploadManager.Attach(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
	suite.ErrorContains(err, "uploadManager.Attach")
}

func (suite *ManagerTestSuite) TestAttach_InvalidPurpose() {
	req := dto.AttachFileRequest{
		ObjectKey: "test",
		Purpose:   suite.purpose,
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{}, upload.ErrPurposeNotFound).Once()

	response, err := suite.uploadManager.Attach(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidPurpose)
	suite.ErrorContains(err, "uploadManager.Attach")
}

func (suite *ManagerTestSuite) TestAttach_FileNotFound() {
	req := dto.AttachFileRequest{
		UploadID:  suite.uploadID,
		ObjectKey: "products/123/img.png",
		Entity:    dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		Purpose:   suite.purpose,
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{}, nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(nil, storage.ErrNotFound).Once()

	response, err := suite.uploadManager.Attach(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrFileNotFound)
	suite.ErrorContains(err, "uploadManager.Attach")
}

func (suite *ManagerTestSuite) TestAttach_FileAlreadyUploaded() {
	req := dto.AttachFileRequest{
		ObjectKey: "key",
		Entity:    dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		UploadID:  suite.uploadID,
		Purpose:   suite.purpose,
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(models.EntityTypeProduct),
		"Purpose":     req.Purpose,
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{}, nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().ExistsByObjectKey(suite.ctx, req.ObjectKey).
		Return(true, nil).Once()

	response, err := suite.uploadManager.Attach(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrFileAlreadyUploaded)
	suite.ErrorContains(err, "uploadManager.Attach")
}

func (suite *ManagerTestSuite) TestAttach_InvalidDetectedContentType() {
	req := dto.AttachFileRequest{
		ObjectKey: "key",
		Entity:    dto.NewUploadEntity(suite.entityID, string(models.EntityTypeProduct)),
		UploadID:  suite.uploadID,
		Purpose:   suite.purpose,
	}

	filePolicy := upload.Policy{
		MaxSize: 5 << 20,
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
		},
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(models.EntityTypeProduct),
		"Purpose":     req.Purpose,
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.registry.EXPECT().Get(upload.Purpose(req.Purpose)).
		Return(&upload.Registration{
			Policy:  filePolicy,
			Handler: suite.handler,
		}, nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().ExistsByObjectKey(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.EXPECT().Get(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	suite.ctDetector.EXPECT().Detect(obj).
		Return("application/pdf", nil).Once()

	suite.storage.EXPECT().Delete(suite.ctx, req.ObjectKey).
		Return(nil).Once()

	response, err := suite.uploadManager.Attach(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidFileType)
	suite.ErrorContains(err, "uploadManager.Attach")
}
