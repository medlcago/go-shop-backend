package service

import (
	"context"
	"errors"
	"go-shop-backend/config"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	repoMocks "go-shop-backend/internal/repository/mocks"
	serviceMocks "go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/apperrors"
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

type UploadServiceTestSuite struct {
	suite.Suite
	storage       *storageMocks.MockStorage
	entityService *serviceMocks.MockEntityService
	uploadRepo    *repoMocks.MockUploadRepository
	uploadConfig  config.Upload
	ctDetector    *contenttypeMocks.MockDetector
	uploadService *uploadService

	ctx      context.Context
	uploadID uuid.UUID
	entityID uuid.UUID
}

func (suite *UploadServiceTestSuite) SetupTest() {
	suite.storage = storageMocks.NewMockStorage(suite.T())
	suite.entityService = serviceMocks.NewMockEntityService(suite.T())
	suite.uploadRepo = repoMocks.NewMockUploadRepository(suite.T())
	suite.uploadConfig = config.Upload{
		MaxFileSize:     1024 * 1024 * 5,
		PresignedUrlTTL: time.Minute,
	}
	suite.ctDetector = contenttypeMocks.NewMockDetector(suite.T())
	suite.uploadService = NewUploadService(
		suite.storage,
		suite.entityService,
		suite.uploadRepo,
		suite.uploadConfig,
		suite.ctDetector,
	)

	suite.ctx = context.Background()
	suite.uploadID = uuid.New()
	suite.entityID = uuid.New()
}

func TestUploadServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UploadServiceTestSuite))
}

// ==================== SignURL Tests ====================

func (suite *UploadServiceTestSuite) TestSignURL_Success() {
	req := dto.SignURLRequest{
		ContentType: "image/png",
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
		Ext: "png",
	}

	presignedPost := &storage.PresignedPost{
		URL: "https://s3.example.com/media",
		Fields: map[string]string{
			"Entity-Type": string(req.Entity.Type),
			"Entity-Id":   req.Entity.ID.String(),
		},
	}

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	suite.storage.EXPECT().CreatePresignedPost(suite.ctx, mock.MatchedBy(func(opts storage.PresignedPostOptions) bool {
		return opts.ContentType == "image/png" &&
			opts.Metadata["Entity-Id"] == suite.entityID.String() &&
			opts.Metadata["Entity-Type"] == string(dto.EntityProduct)
	})).Return(presignedPost, nil)

	resp, err := suite.uploadService.SignURL(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(resp)

	suite.Equal(presignedPost.URL, resp.UploadURL)
	suite.Equal("image/png", resp.ContentType)
}

func (suite *UploadServiceTestSuite) TestSignURL_InvalidExtOrContentType() {
	tests := []struct {
		name        string
		ext         string
		contentType string
		expectedErr error
	}{
		{
			name:        "invalid content type",
			ext:         "jpg",
			contentType: "application/pdf",
			expectedErr: apperrors.ErrInvalidImageFormat,
		},
		{
			name:        "mismatch ext and content type",
			ext:         "png",
			contentType: "image/jpeg",
			expectedErr: apperrors.ErrContentTypeMismatch,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := dto.SignURLRequest{
				Ext:         tt.ext,
				ContentType: tt.contentType,
				Entity: dto.Entity{
					Type: dto.EntityProduct,
					ID:   uuid.New(),
				},
			}

			resp, err := suite.uploadService.SignURL(suite.ctx, req)

			suite.Nil(resp)
			suite.ErrorIs(err, tt.expectedErr)
		})
	}
}

func (suite *UploadServiceTestSuite) TestSignURL_EntityNotFound() {
	req := dto.SignURLRequest{
		ContentType: "image/jpeg",
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
		Ext: "jpg",
	}

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(apperrors.ErrEntityNotFound).Once()

	resp, err := suite.uploadService.SignURL(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrEntityNotFound)
}

func (suite *UploadServiceTestSuite) TestSignURL_UnknownEntityType() {
	req := dto.SignURLRequest{
		ContentType: "image/jpeg",
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityType("test"),
		},
		Ext: "jpg",
	}

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(apperrors.ErrUnknownEntityType).Once()

	resp, err := suite.uploadService.SignURL(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrUnknownEntityType)
}

func (suite *UploadServiceTestSuite) TestSignURL_EntityServiceError() {
	req := dto.SignURLRequest{
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
		Ext:         "jpg",
		ContentType: "image/jpeg",
	}

	expectedErr := errors.New("database error")
	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(expectedErr)

	resp, err := suite.uploadService.SignURL(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorContains(err, expectedErr.Error())
}

func (suite *UploadServiceTestSuite) TestSignURL_StorageError() {
	req := dto.SignURLRequest{
		Ext:         "jpg",
		ContentType: "image/jpeg",
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
	}

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	expectedErr := errors.New("storage error")
	suite.storage.EXPECT().CreatePresignedPost(suite.ctx, mock.Anything).
		Return(nil, expectedErr).Once()

	resp, err := suite.uploadService.SignURL(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorContains(err, expectedErr.Error())
}

// ==================== Create Tests ====================

func (suite *UploadServiceTestSuite) TestSave_Success() {
	req := dto.UploadRequest{
		UploadID:  suite.uploadID,
		ObjectKey: "products/123/img.jpg",
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objectInfo := &storage.ObjectInfo{
		Size:     1024,
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objectInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.EXPECT().Open(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	detectedCT := "image/jpeg"
	suite.ctDetector.EXPECT().Detect(obj).
		Return(detectedCT, nil).Once()

	suite.uploadRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(u *models.Upload) bool {
		return u.ObjectKey == req.ObjectKey &&
			u.EntityID == suite.entityID &&
			*u.ContentType == detectedCT
	})).Return(nil).Once()

	url := "https://s3.example.com/img.jpg"
	suite.storage.EXPECT().PublicURL(suite.ctx, req.ObjectKey).
		Return(url).Once()

	resp, err := suite.uploadService.Save(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(resp)

	suite.Equal(url, resp.URL)
	suite.Equal(detectedCT, *resp.ContentType)
}

func (suite *UploadServiceTestSuite) TestSave_EntityNotFound() {
	req := dto.UploadRequest{
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
	}

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(apperrors.ErrEntityNotFound).Once()

	resp, err := suite.uploadService.Save(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrEntityNotFound)
}

func (suite *UploadServiceTestSuite) TestSave_InvalidMetadata() {
	req := dto.UploadRequest{
		ObjectKey: "key",
		UploadID:  suite.uploadID,
		Entity: dto.Entity{
			Type: dto.EntityProduct,
			ID:   suite.entityID,
		},
	}

	tests := []struct {
		name        string
		metadata    map[string]string
		expectedErr error
	}{
		{
			name: "invalid upload id",
			metadata: map[string]string{
				"Upload-Id":   uuid.NewString(),
				"Entity-Id":   suite.entityID.String(),
				"Entity-Type": string(dto.EntityProduct),
			},
			expectedErr: apperrors.ErrInvalidUploadID,
		},
		{
			name: "invalid entity id",
			metadata: map[string]string{
				"Upload-Id":   suite.uploadID.String(),
				"Entity-Id":   uuid.NewString(),
				"Entity-Type": string(dto.EntityProduct),
			},
			expectedErr: apperrors.ErrInvalidEntityID,
		},
		{
			name: "invalid entity type",
			metadata: map[string]string{
				"Upload-Id":   suite.uploadID.String(),
				"Entity-Id":   suite.entityID.String(),
				"Entity-Type": "test",
			},
			expectedErr: apperrors.ErrInvalidEntityID,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			objInfo := &storage.ObjectInfo{
				Size:     1024,
				Metadata: tt.metadata,
			}
			suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
				Return(nil).Once()

			suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
				Return(objInfo, nil).Once()

			resp, err := suite.uploadService.Save(suite.ctx, req)

			suite.Nil(resp)
			suite.ErrorIs(err, tt.expectedErr)
		})
	}
}

func (suite *UploadServiceTestSuite) TestSave_FileAlreadyUploaded() {
	req := dto.UploadRequest{
		ObjectKey: "key",
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
		UploadID: suite.uploadID,
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(true, nil).Once()

	resp, err := suite.uploadService.Save(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrFileAlreadyUploaded)
}

func (suite *UploadServiceTestSuite) TestSave_InvalidDetectedContentType() {
	req := dto.UploadRequest{
		ObjectKey: "key",
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
		UploadID: suite.uploadID,
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.EXPECT().Open(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	suite.ctDetector.EXPECT().Detect(obj).
		Return("application/pdf", nil).Once()

	resp, err := suite.uploadService.Save(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrInvalidImageFormat)
}

func (suite *UploadServiceTestSuite) TestSave_RepositoryError() {
	req := dto.UploadRequest{
		ObjectKey: "key",
		UploadID:  suite.uploadID,
		Entity: dto.Entity{
			ID:   suite.entityID,
			Type: dto.EntityProduct,
		},
	}

	metadata := map[string]string{
		"Upload-Id":   suite.uploadID.String(),
		"Entity-Id":   suite.entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	obj := testutils.NewReadSeekCloser()

	suite.entityService.EXPECT().Exists(suite.ctx, req.Entity.Type, req.Entity.ID).
		Return(nil).Once()

	suite.storage.EXPECT().GetObjectInfo(suite.ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.EXPECT().Exists(suite.ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.EXPECT().Open(suite.ctx, req.ObjectKey).
		Return(obj, nil).Once()

	suite.ctDetector.EXPECT().Detect(obj).
		Return("image/jpeg", nil).Once()

	expectedErr := errors.New("database save error")
	suite.uploadRepo.EXPECT().Create(suite.ctx, mock.Anything).
		Return(expectedErr).Once()

	resp, err := suite.uploadService.Save(suite.ctx, req)

	suite.Nil(resp)
	suite.ErrorContains(err, expectedErr.Error())
}
