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
	"go-shop-backend/pkg/storage"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockContentTypeDetector struct {
	mock.Mock
}

func (m *MockContentTypeDetector) Detect(r io.ReadSeeker) (string, error) {
	args := m.Called(r)
	return args.String(0), args.Error(1)
}

type UploadServiceTestSuite struct {
	suite.Suite
	storage       *storage.MockStorage
	entityService *serviceMocks.EntityServiceMock
	uploadRepo    *repoMocks.UploadRepositoryMock
	uploadConfig  config.Upload
	ctDetector    *MockContentTypeDetector
	uploadService UploadService
}

func (suite *UploadServiceTestSuite) SetupTest() {
	suite.storage = new(storage.MockStorage)
	suite.entityService = new(serviceMocks.EntityServiceMock)
	suite.uploadRepo = new(repoMocks.UploadRepositoryMock)
	suite.uploadConfig = config.Upload{
		MaxFileSize:     1024 * 1024 * 5,
		PresignedUrlTTL: time.Minute,
	}
	suite.ctDetector = new(MockContentTypeDetector)
	suite.uploadService = NewUploadService(
		suite.storage,
		suite.entityService,
		suite.uploadRepo,
		suite.uploadConfig,
		suite.ctDetector,
	)
}

func TestUploadServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UploadServiceTestSuite))
}

// ==================== SignURL Tests ====================

func (suite *UploadServiceTestSuite) TestSignURL_Success() {
	ctx := context.Background()
	entityID := uuid.New()
	req := dto.SignURLRequest{
		ContentType: "image/png",
		Entity: dto.Entity{
			ID:   entityID,
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

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(true, nil).Once()

	suite.storage.On("CreatePresignedPost", mock.Anything, mock.MatchedBy(func(opts storage.PresignedPostOptions) bool {
		return opts.ContentType == "image/png" &&
			opts.Metadata["Entity-Id"] == entityID.String() &&
			opts.Metadata["Entity-Type"] == string(dto.EntityProduct)
	})).Return(presignedPost, nil)

	resp, err := suite.uploadService.SignURL(ctx, req)

	suite.NoError(err)
	suite.NotNil(resp)

	suite.Equal(presignedPost.URL, resp.UploadURL)
	suite.Equal("image/png", resp.ContentType)
}

func (suite *UploadServiceTestSuite) TestSignURL_InvalidExtOrContentType() {
	ctx := context.Background()
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

			resp, err := suite.uploadService.SignURL(ctx, req)

			suite.Nil(resp)
			suite.ErrorIs(err, tt.expectedErr)
		})
	}
}

func (suite *UploadServiceTestSuite) TestSignURL_EntityNotFound() {
	ctx := context.Background()
	entityID := uuid.New()
	req := dto.SignURLRequest{
		ContentType: "image/jpeg",
		Entity: dto.Entity{
			ID:   entityID,
			Type: dto.EntityProduct,
		},
		Ext: "jpg",
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(false, nil).Once()

	resp, err := suite.uploadService.SignURL(ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrEntityNotFound)
}

func (suite *UploadServiceTestSuite) TestSignURL_UnknownEntityType() {
	ctx := context.Background()
	entityID := uuid.New()
	req := dto.SignURLRequest{
		ContentType: "image/jpeg",
		Entity: dto.Entity{
			ID:   entityID,
			Type: dto.EntityType("test"),
		},
		Ext: "jpg",
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(false, apperrors.ErrUnknownEntityType).Once()

	resp, err := suite.uploadService.SignURL(ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrUnknownEntityType)
}

func (suite *UploadServiceTestSuite) TestSignURL_EntityServiceError() {
	ctx := context.Background()
	req := dto.SignURLRequest{
		Entity: dto.Entity{
			ID:   uuid.New(),
			Type: dto.EntityProduct,
		},
		Ext:         "jpg",
		ContentType: "image/jpeg",
	}

	expectedErr := errors.New("database error")
	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(false, errors.New("database error"))

	resp, err := suite.uploadService.SignURL(ctx, req)

	suite.Nil(resp)
	suite.ErrorContains(err, expectedErr.Error())
}

func (suite *UploadServiceTestSuite) TestSignURL_StorageError() {
	ctx := context.Background()
	req := dto.SignURLRequest{
		Ext:         "jpg",
		ContentType: "image/jpeg",
		Entity: dto.Entity{
			ID:   uuid.New(),
			Type: dto.EntityProduct,
		},
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(true, nil).Once()

	expectedErr := errors.New("storage error")
	suite.storage.On("CreatePresignedPost", ctx, mock.Anything).
		Return(&storage.PresignedPost{}, expectedErr).Once()

	resp, err := suite.uploadService.SignURL(ctx, req)

	suite.Nil(resp)
	suite.ErrorContains(err, expectedErr.Error())
}

// ==================== Save Tests ====================

func (suite *UploadServiceTestSuite) TestSave_Success() {
	ctx := context.Background()
	entityID := uuid.New()
	uploadID := uuid.New()
	req := dto.UploadRequest{
		UploadID:  uploadID,
		ObjectKey: "products/123/img.jpg",
		Entity: dto.Entity{
			ID:   entityID,
			Type: dto.EntityProduct,
		},
	}

	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Id":   entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objectInfo := &storage.ObjectInfo{
		Size:     1024,
		Metadata: metadata,
	}

	reader := strings.NewReader("test")
	obj := struct {
		io.ReadSeeker
		io.Closer
	}{
		reader,
		io.NopCloser(reader),
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(true, nil).Once()

	suite.storage.On("GetObjectInfo", ctx, req.ObjectKey).
		Return(objectInfo, nil).Once()

	suite.uploadRepo.On("Exists", ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.On("Open", ctx, req.ObjectKey).
		Return(obj, nil).Once()

	detectedCT := "image/jpeg"
	suite.ctDetector.On("Detect", obj).
		Return(detectedCT, nil).Once()

	suite.uploadRepo.On("Save", ctx, mock.MatchedBy(func(u *models.Upload) bool {
		return u.ObjectKey == req.ObjectKey &&
			u.EntityID == entityID &&
			*u.ContentType == detectedCT
	})).Return(nil).Once()

	url := "https://s3.example.com/img.jpg"
	suite.storage.On("PublicURL", ctx, req.ObjectKey).Return(url).Once()

	resp, err := suite.uploadService.Save(ctx, req)

	suite.NoError(err)
	suite.NotNil(resp)

	suite.Equal(url, resp.URL)
	suite.Equal(detectedCT, *resp.ContentType)
}

func (suite *UploadServiceTestSuite) TestSave_EntityNotFound() {
	ctx := context.Background()
	req := dto.UploadRequest{
		Entity: dto.Entity{
			ID:   uuid.New(),
			Type: dto.EntityProduct,
		},
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(false, nil).Once()

	resp, err := suite.uploadService.Save(ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrEntityNotFound)
}

func (suite *UploadServiceTestSuite) TestSave_InvalidMetadata() {
	ctx := context.Background()
	uploadID := uuid.New()
	entityID := uuid.New()
	req := dto.UploadRequest{
		ObjectKey: "key",
		UploadID:  uploadID,
		Entity: dto.Entity{
			Type: dto.EntityProduct,
			ID:   entityID,
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
				"Entity-Id":   entityID.String(),
				"Entity-Type": string(dto.EntityProduct),
			},
			expectedErr: apperrors.ErrInvalidUploadID,
		},
		{
			name: "invalid entity id",
			metadata: map[string]string{
				"Upload-Id":   uploadID.String(),
				"Entity-Id":   uuid.NewString(),
				"Entity-Type": string(dto.EntityProduct),
			},
			expectedErr: apperrors.ErrInvalidEntityID,
		},
		{
			name: "invalid entity type",
			metadata: map[string]string{
				"Upload-Id":   uploadID.String(),
				"Entity-Id":   entityID.String(),
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
			suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
				Return(true, nil).Once()

			suite.storage.On("GetObjectInfo", ctx, req.ObjectKey).
				Return(objInfo, nil).Once()

			resp, err := suite.uploadService.Save(ctx, req)

			suite.Nil(resp)
			suite.Error(err)
			suite.ErrorIs(err, tt.expectedErr)
		})
	}
}

func (suite *UploadServiceTestSuite) TestSave_FileAlreadyUploaded() {
	ctx := context.Background()
	uploadID := uuid.New()
	entityID := uuid.New()
	req := dto.UploadRequest{
		ObjectKey: "key",
		Entity: dto.Entity{
			ID:   entityID,
			Type: dto.EntityProduct,
		},
		UploadID: uploadID,
	}

	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Id":   entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(true, nil).Once()

	suite.storage.On("GetObjectInfo", ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.On("Exists", ctx, req.ObjectKey).
		Return(true, nil).Once()

	resp, err := suite.uploadService.Save(ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrFileAlreadyUploaded)
}

func (suite *UploadServiceTestSuite) TestSave_InvalidDetectedContentType() {
	ctx := context.Background()
	uploadID := uuid.New()
	entityID := uuid.New()
	req := dto.UploadRequest{
		ObjectKey: "key",
		Entity: dto.Entity{
			ID:   entityID,
			Type: dto.EntityProduct,
		},
		UploadID: uploadID,
	}

	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Id":   entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	reader := strings.NewReader("test")
	obj := struct {
		io.ReadSeeker
		io.Closer
	}{
		reader,
		io.NopCloser(reader),
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(true, nil).Once()

	suite.storage.On("GetObjectInfo", ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.On("Exists", ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.On("Open", ctx, req.ObjectKey).
		Return(obj, nil).Once()

	suite.ctDetector.On("Detect", obj).
		Return("application/pdf", nil).Once()

	resp, err := suite.uploadService.Save(ctx, req)

	suite.Nil(resp)
	suite.ErrorIs(err, apperrors.ErrInvalidImageFormat)
}

func (suite *UploadServiceTestSuite) TestSave_RepositoryError() {
	ctx := context.Background()
	uploadID := uuid.New()
	entityID := uuid.New()

	req := dto.UploadRequest{
		ObjectKey: "key",
		UploadID:  uploadID,
		Entity: dto.Entity{
			ID:   entityID,
			Type: dto.EntityProduct,
		},
	}

	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Id":   entityID.String(),
		"Entity-Type": string(dto.EntityProduct),
	}

	objInfo := &storage.ObjectInfo{
		Metadata: metadata,
	}

	reader := strings.NewReader("test")
	obj := struct {
		io.ReadSeeker
		io.Closer
	}{
		reader,
		io.NopCloser(reader),
	}

	suite.entityService.On("Exists", ctx, req.Entity.Type, req.Entity.ID).
		Return(true, nil).Once()

	suite.storage.On("GetObjectInfo", ctx, req.ObjectKey).
		Return(objInfo, nil).Once()

	suite.uploadRepo.On("Exists", ctx, req.ObjectKey).
		Return(false, nil).Once()

	suite.storage.On("Open", ctx, req.ObjectKey).
		Return(obj, nil).Once()

	suite.ctDetector.On("Detect", obj).
		Return("image/jpeg", nil).Once()

	expectedErr := errors.New("database save error")
	suite.uploadRepo.On("Save", ctx, mock.Anything).
		Return(expectedErr).Once()

	resp, err := suite.uploadService.Save(ctx, req)

	suite.Nil(resp)
	suite.ErrorContains(err, expectedErr.Error())
}
