package v1

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andreyxaxa/Image-Processor/internal/controller/restapi/v1/response"
	"github.com/andreyxaxa/Image-Processor/internal/controller/restapi/v1/validate"
	"github.com/andreyxaxa/Image-Processor/internal/dto"
	"github.com/andreyxaxa/Image-Processor/pkg/types/errs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// @Summary  	Upload and process image
// @Description Uploads image to S3, save metadata to postgres, save metadata to outbox(postgres)
// @Tags 		images
// @Accept 		mpfd
// @Produce 	json
// @Param 		file 	  formData file   true  "Image file(jpg, png, gif)"
// @Param 		operation formData string true  "Operation" Enums(resize, thumbnail, watermark)
// @Param 		width 	  formData int    false "Width(required for resize operation)"
// @Param 		height 	  formData int 	  false "Height(required for resize operation)"
// @Success 	201 {object} response.ProcessImage
// @Failure 	400 {object} response.Error "Empty file or wrong parameters"
// @Failure 	413 {object} response.Error "File too large"
// @Failure 	415 {object} response.Error "Unsupported format"
// @Failure 	500 {object} response.Error "Internal"
// @Router 		/v1/upload [post]
func (r *V1) processImage(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "file is required")
	}

	// 1. валидация размера
	if file.Size == 0 {
		return errorResponse(ctx, http.StatusBadRequest, "file is empty")
	}

	if file.Size > validate.MaxFileSize {
		return errorResponse(ctx, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("file size cant be more than %d bytes", validate.MaxFileSize))
	}

	// 2. валидация content type
	contentType := file.Header.Get("Content-Type")
	if !validate.AllowedContentTypes[contentType] {
		return errorResponse(ctx, http.StatusUnsupportedMediaType, "unsupported file type. Allowed: jpeg, png, gif")
	}

	// 3. валидация расширения
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !validate.AllowedExtensions[ext] {
		return errorResponse(ctx, http.StatusUnsupportedMediaType, "unsupported file extension. Allowed: .jpg, .jpeg, .png, .gif")
	}

	// 4. валидация операции
	operation := strings.ToLower(ctx.FormValue("operation"))
	if operation == "" {
		return errorResponse(ctx, http.StatusBadRequest, "operation is required")
	}

	var op dto.Operation

	switch operation {
	case "resize":
		// width
		widthStr := ctx.FormValue("width")
		if widthStr == "" {
			return errorResponse(ctx, http.StatusBadRequest, "width is required for resize")
		}
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, "width must be a number")
		}
		if width < validate.MinResizeWidth || width > validate.MaxResizeWidth {
			return errorResponse(ctx, http.StatusBadRequest,
				fmt.Sprintf("width must be between %d and %d", validate.MinResizeWidth, validate.MaxResizeWidth))
		}

		// height
		heightStr := ctx.FormValue("height")
		if heightStr == "" {
			return errorResponse(ctx, http.StatusBadRequest, "height is required for resize")
		}
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, "height must be a number")
		}
		if height < validate.MinResizeHeight || height > validate.MaxResizeHeight {
			return errorResponse(ctx, http.StatusBadRequest,
				fmt.Sprintf("height must be between %d and %d", validate.MinResizeHeight, validate.MaxResizeHeight))
		}

		op = dto.Operation{
			Operation: "resize",
			Width:     &width,
			Height:    &height,
		}
	case "thumbnail":
		op = dto.Operation{
			Operation: "thumbnail",
		}
	case "watermark":
		// text
		textStr := ctx.FormValue("text")
		if textStr == "" {
			return errorResponse(ctx, http.StatusBadRequest, "text is required for watermark")
		}

		if len(textStr) < validate.MinTextLen || len(textStr) > validate.MaxTextLen {
			return errorResponse(ctx, http.StatusBadRequest,
				fmt.Sprintf("text length must be between %d and %d", validate.MinTextLen, validate.MaxTextLen))
		}

		op = dto.Operation{
			Operation: "watermark",
		}
	default:
		return errorResponse(ctx, http.StatusBadRequest, "invalid operation. Allowed: resize, thumbnail, watermark")
	}

	// 5. открытие файла
	fileReader, err := file.Open()
	if err != nil {
		r.logger.Error(err, "restapi - v1 - processImage")

		return errorResponse(ctx, http.StatusInternalServerError, "problems with opening the file")
	}
	defer fileReader.Close()

	// 6. загружаем
	image, err := r.img.UploadNewImage(ctx.UserContext(), fileReader, file.Filename, contentType, file.Size, op)
	if err != nil {
		r.logger.Error(err, "restapi - v1 - processImage")

		return errorResponse(ctx, http.StatusInternalServerError, "storage problems")
	}

	// 7. ответ
	resp := response.ProcessImage{
		ImageID:      image.ID.String(),
		OriginalName: image.OriginalName,
		Size:         int(image.Size),
		ContentType:  image.ContentType,
		Status:       string(image.Status),
		Operation:    op.Operation,
		CreatedAt:    image.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return ctx.Status(http.StatusCreated).JSON(resp)
}

// @Summary 	Get processed image
// @Description Downloads processed image from S3 by key
// @Tags 		images
// @Produce 	image/jpeg,image/png,image/gif
// @Param 		id path string true "Image ID(uuid)"
// @Success 	200 {file} 	binary
// @Failure 	400 {object} response.Error "Invalid ID"
// @Failure 	404 {object} response.Error "Image not found"
// @Failure 	500 {object} response.Error "Internal"
// @Router 		/v1/image/{id} [get]
func (r *V1) getProcessedImage(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")

	if idStr == "" {
		return errorResponse(ctx, http.StatusBadRequest, "invalid id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid id")
	}

	processedKey, contentType, err := r.img.GetProcessedKeyByID(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "image not found")
		}
		r.logger.Error(err, "restapi - v1 - getProcessedImage")

		return errorResponse(ctx, http.StatusInternalServerError, "storage problems")
	}

	body, err := r.img.DownloadImage(ctx.UserContext(), processedKey)
	if err != nil {
		return errorResponse(ctx, http.StatusInternalServerError, "storage problems")
	}

	ctx.Set(fiber.HeaderContentType, contentType)

	return ctx.SendStream(body)
}

// @Summary 	Delete image
// @Description Deletes image from all storages(S3, postgres(main table + outbox(cascade)))
// @Tags 		images
// @Param		id 	path	 string true "Image ID(uuid)"
// @Success		204 "Deleted"
// @Failure 	400 {object} response.Error "Invalid ID"
// @Failure 	404 {object} response.Error "Image not found"
// @Failure 	500 {object} response.Error "Internal"
// @Router 		/v1/image/{id} [delete]
func (r *V1) deleteImage(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")

	if idStr == "" {
		return errorResponse(ctx, http.StatusBadRequest, "invalid id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid id")
	}

	err = r.img.DeleteImage(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "image not found")
		}
		r.logger.Error(err, "restapi - v1 - deleteImage")

		return errorResponse(ctx, http.StatusInternalServerError, "problem storage")
	}

	return ctx.SendStatus(http.StatusNoContent)
}
