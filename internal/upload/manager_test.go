package upload_test

import (
	"context"
	"errors"
	"go-shop-backend/config"
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
	storage        *storageMocks.MockStorage
	uploadRepo     *repoMocks.MockUploadRepository
	uploadConfig   config.Upload
	ctDetector     *contenttypeMocks.MockDetector
	policyProvider *uploadMocks.MockPolicyProvider
	uploadManager  upload.Manager

	ctx      context.Context
	uploadID uuid.UUID
	entityID uuid.UUID
}

func (suite *ManagerTestSuite) SetupTest() {
	suite.storage = storageMocks.NewMockStorage(suite.T())
	suite.uploadRepo = repoMocks.NewMockUploadRepository(suite.T())
	suite.uploadConfig = config.Upload{
		MaxFileSize:     1024 * 1024 * 5,
		PresignedUrlTTL: time.Minute,
	}
	suite.ctDetector = contenttypeMocks.NewMockDetector(suite.T())
	suite.policyProvider = uploadMocks.NewMockPolicyProvider(suite.T())
	suite.uploadManager = upload.NewManager(
		suite.storage,
		suite.uploadRepo,
		suite.uploadConfig,
		suite.ctDetector,
		suite.policyProvider,
		testutils.NewLogger(),
	)

	suite.ctx = context.Background()
	suite.uploadID = uuid.New()
	suite.entityID = uuid.New()
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

// ==================== SignURL Tests ====================

func (suite *ManagerTestSuite) TestSignURL_Success() {
	req := upload.SignURLRequest{
		ContentType: "image/png",
		Entity:      upload.NewProductEntity(suite.entityID),
		Ext:         "png",
	}

	presignedPost := &storage.TemporaryUploadURL{
		URL: "https://s3.example.com/media",
		Fields: map[string]string{
			"Entity-Type": string(req.Entity.Type),
			"Entity-Id":   req.Entity.ID.String(),
		},
	}

	constrains := upload.FileConstraints{
		MaxSize: 1 << 20,
		AllowedFormats: []upload.Format{
			{
				Extensions: []string{"png"}, ContentType: "image/png",
			},
		},
	}

	suite.policyProvider.EXPECT().Get(upload.ProductImagePolicy).
		Return(constrains, nil).Once()

	suite.storage.EXPECT().TemporaryUploadURL(suite.ctx, mock.MatchedBy(func(opts storage.TemporaryUploadURLOptions) bool {
		return opts.ContentType == "image/png" &&
			opts.Metadata["Entity-Id"] == suite.entityID.String() &&
			opts.Metadata["Entity-Type"] == string(upload.EntityTypeProduct) &&
			opts.MaxSize == 1<<20 &&
			!opts.Expires.IsZero()
	})).Return(presignedPost, nil).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, req, upload.ProductImagePolicy)

	suite.NoError(err)
	suite.NotNil(response)

	suite.Equal(presignedPost.URL, response.UploadURL)
	suite.Equal("image/png", response.ContentType)
}

func (suite *ManagerTestSuite) TestSignURL_ContentTypeMismatch() {
	constraints := upload.FileConstraints{
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
			{Extensions: []string{"jpg", "jpeg"}, ContentType: "image/jpeg"},
		},
	}

	suite.policyProvider.EXPECT().Get(upload.ProductImagePolicy).
		Return(constraints, nil).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, upload.SignURLRequest{
		ContentType: "video/mp4",
		Ext:         "mp4",
	}, upload.ProductImagePolicy)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrContentTypeMismatch)

}

func (suite *ManagerTestSuite) TestSignURL_UnknownPolicy() {
	policyErr := errors.New("unknown policy")
	suite.policyProvider.EXPECT().Get(upload.Policy("test")).
		Return(upload.FileConstraints{}, policyErr).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, upload.SignURLRequest{}, "test")

	suite.Nil(response)
	suite.ErrorContains(err, policyErr.Error())
}

func (suite *ManagerTestSuite) TestSignURL_StorageError() {
	req := upload.SignURLRequest{
		Ext:         "png",
		ContentType: "image/png",
		Entity:      upload.NewProductEntity(suite.entityID),
	}

	constrains := upload.FileConstraints{
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
		},
	}

	suite.policyProvider.EXPECT().Get(upload.ProductImagePolicy).
		Return(constrains, nil).Once()

	expectedErr := errors.New("storage error")
	suite.storage.EXPECT().TemporaryUploadURL(suite.ctx, mock.Anything).
		Return(nil, expectedErr).Once()

	response, err := suite.uploadManager.SignURL(suite.ctx, req, upload.ProductImagePolicy)

	suite.Nil(response)
	suite.ErrorContains(err, expectedErr.Error())
}

// ==================== Save Tests ====================

func (suite *ManagerTestSuite) TestSave_Success() {
	req := upload.SaveUploadRequest{
		UploadID:  suite.uploadID,
		ObjectKey: "products/123/img.png",
		Entity:    upload.NewProductEntity(suite.entityID),
	}

	constrains := upload.FileConstraints{
		MaxSize: 5 << 20,
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
		},
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(upload.EntityTypeProduct),
	}

	objectInfo := &storage.ObjectInfo{
		Size:     1024,
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objectInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.policyProvider.EXPECT().Get(upload.ProductImagePolicy).
		Return(constrains, nil).Once()

	suite.storage.EXPECT().Open(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	detectedCT := "image/png"
	suite.ctDetector.EXPECT().Detect(obj).
		Return(detectedCT, nil).Once()

	suite.uploadRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(u *models.Upload) bool {
		return u.ObjectKey == req.ObjectKey &&
			u.EntityID == suite.entityID &&
			u.EntityType == string(upload.EntityTypeProduct) &&
			*u.ContentType == detectedCT
	})).Return(nil).Once()

	url := "https://s3.example.com/img.png"
	suite.storage.EXPECT().PublicURL(req.ObjectKey).
		Return(url).Once()

	response, err := suite.uploadManager.Save(suite.ctx, req, upload.ProductImagePolicy)

	suite.NoError(err)
	suite.NotNil(response)

	suite.Equal(url, response.URL)
	suite.Equal(detectedCT, *response.ContentType)
}

func (suite *ManagerTestSuite) TestSave_NotFound() {
	req := upload.SaveUploadRequest{
		UploadID:  suite.uploadID,
		ObjectKey: "products/123/img.png",
		Entity:    upload.NewProductEntity(suite.entityID),
	}

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(nil, errors.New("error")).Once()

	response, err := suite.uploadManager.Save(suite.ctx, req, upload.ProductImagePolicy)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrNotFound)
}

func (suite *ManagerTestSuite) TestSave_FileAlreadyUploaded() {
	req := upload.SaveUploadRequest{
		ObjectKey: "key",
		Entity:    upload.NewProductEntity(suite.entityID),
		UploadID:  suite.uploadID,
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(upload.EntityTypeProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(true, nil).Once()

	response, err := suite.uploadManager.Save(suite.ctx, req, upload.ProductImagePolicy)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrFileAlreadyUploaded)
}

func (suite *ManagerTestSuite) TestSave_InvalidDetectedContentType() {
	req := upload.SaveUploadRequest{
		ObjectKey: "key",
		Entity:    upload.NewProductEntity(suite.entityID),
		UploadID:  suite.uploadID,
	}

	constrains := upload.FileConstraints{
		MaxSize: 5 << 20,
		AllowedFormats: []upload.Format{
			{Extensions: []string{"png"}, ContentType: "image/png"},
		},
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(upload.EntityTypeProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.policyProvider.EXPECT().Get(upload.ProductImagePolicy).
		Return(constrains, nil).Once()

	suite.storage.EXPECT().Open(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	suite.ctDetector.EXPECT().Detect(obj).
		Return("application/pdf", nil).Once()

	suite.storage.EXPECT().Delete(suite.ctx, req.ObjectKey).
		Return(nil).Once()

	response, err := suite.uploadManager.Save(suite.ctx, req, upload.ProductImagePolicy)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidFileType)
}

func (suite *ManagerTestSuite) TestSave_RepositoryError() {
	req := upload.SaveUploadRequest{
		ObjectKey: "key",
		UploadID:  suite.uploadID,
		Entity:    upload.NewProductEntity(suite.entityID),
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(upload.EntityTypeProduct),
	}

	constrains := upload.FileConstraints{
		MaxSize: 5 << 20,
		AllowedFormats: []upload.Format{
			{Extensions: []string{"jpg"}, ContentType: "image/jpeg"},
		},
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.policyProvider.EXPECT().Get(upload.ProductImagePolicy).
		Return(constrains, nil).Once()

	suite.storage.EXPECT().Open(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	suite.ctDetector.EXPECT().Detect(obj).
		Return("image/jpeg", nil).Once()

	expectedErr := errors.New("create error")
	suite.uploadRepo.EXPECT().Create(suite.ctx, mock.Anything).
		Return(expectedErr).Once()

	suite.storage.EXPECT().Delete(suite.ctx, req.ObjectKey).
		Return(nil).Once()

	response, err := suite.uploadManager.Save(suite.ctx, req, upload.ProductImagePolicy)

	suite.Nil(response)
	suite.ErrorContains(err, expectedErr.Error())
}
