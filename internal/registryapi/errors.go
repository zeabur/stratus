package registryapi

import "github.com/gofiber/fiber/v3"

type OciErrorCode string

const (
	OciErrorCodeBlobUnknown         OciErrorCode = "BLOB_UNKNOWN"
	OciErrorCodeBlobUploadInvalid   OciErrorCode = "BLOB_UPLOAD_INVALID"
	OciErrorCodeBlobUploadUnknown   OciErrorCode = "BLOB_UPLOAD_UNKNOWN"
	OciErrorCodeDigestInvalid       OciErrorCode = "DIGEST_INVALID"
	OciErrorCodeManifestBlobUnknown OciErrorCode = "MANIFEST_BLOB_UNKNOWN"
	OciErrorCodeManifestInvalid     OciErrorCode = "MANIFEST_INVALID"
	OciErrorCodeManifestUnknown     OciErrorCode = "MANIFEST_UNKNOWN"
	OciErrorCodeNameInvalid         OciErrorCode = "NAME_INVALID"
	OciErrorCodeNameUnknown         OciErrorCode = "NAME_UNKNOWN"
	OciErrorCodeSizeInvalid         OciErrorCode = "SIZE_INVALID"
	OciErrorCodeUnauthorized        OciErrorCode = "UNAUTHORIZED"
	OciErrorCodeDenied              OciErrorCode = "DENIED"
	OciErrorCodeUnsupported         OciErrorCode = "UNSUPPORTED"
	OciErrorCodeTooManyRequests     OciErrorCode = "TOOMANYREQUESTS"

	// Custom
	OciErrorCodeInternalError OciErrorCode = "INTERNAL_ERROR"
)

var ociHTTPStatusCode = map[OciErrorCode]int{
	OciErrorCodeBlobUnknown:         fiber.StatusNotFound,
	OciErrorCodeBlobUploadInvalid:   fiber.StatusBadRequest,
	OciErrorCodeBlobUploadUnknown:   fiber.StatusNotFound,
	OciErrorCodeDigestInvalid:       fiber.StatusBadRequest,
	OciErrorCodeManifestBlobUnknown: fiber.StatusNotFound,
	OciErrorCodeManifestInvalid:     fiber.StatusBadRequest,
	OciErrorCodeManifestUnknown:     fiber.StatusNotFound,
	OciErrorCodeNameInvalid:         fiber.StatusBadRequest,
	OciErrorCodeNameUnknown:         fiber.StatusNotFound,
	OciErrorCodeSizeInvalid:         fiber.StatusBadRequest,
	OciErrorCodeUnauthorized:        fiber.StatusUnauthorized,
	OciErrorCodeDenied:              fiber.StatusForbidden,
	OciErrorCodeUnsupported:         fiber.StatusBadRequest,
	OciErrorCodeTooManyRequests:     fiber.StatusTooManyRequests,
	OciErrorCodeInternalError:       fiber.StatusInternalServerError,
}

type RegistryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Errors []RegistryError `json:"errors"`
}

func errResponse(c fiber.Ctx, code OciErrorCode, message string) error {
	status, ok := ociHTTPStatusCode[code]
	if !ok {
		status = fiber.StatusBadRequest
	}
	return c.Status(status).JSON(&ErrorResponse{
		Errors: []RegistryError{{Code: string(code), Message: message}},
	})
}
