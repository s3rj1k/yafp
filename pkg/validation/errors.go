package validation

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ErrorResponse(err error, fieldNameReplacer *strings.Replacer) string {
	validationErrors := new(validator.ValidationErrors)
	resp := make([]string, 0)

	resp = append(resp, fmt.Sprintf("%d Bad Request", http.StatusBadRequest))

	if errors.As(err, validationErrors) {
		for _, el := range *validationErrors {
			if fieldNameReplacer == nil {
				resp = append(resp,
					fmt.Sprintf(
						"* URL query parameter validation for '%s' failed on the '%s' tag",
						el.Field(),
						el.Tag(),
					),
				)
			} else {
				resp = append(resp,
					fmt.Sprintf(
						"* URL query parameter validation for '%s' failed on the '%s' tag",
						fieldNameReplacer.Replace(el.Field()),
						el.Tag(),
					),
				)
			}
		}
	}

	return strings.Join(resp, "\n")
}
