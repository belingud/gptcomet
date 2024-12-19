from prompt_toolkit.validation import ValidationError, Validator

from gptcomet.utils import is_float

KEYS_VALIDATOR = {
    "retries": {"validator": str.isdecimal, "msg": "`retries` must be a positive integer"},
    "max_tokens": {"validator": str.isdecimal, "msg": "`max_tokens` must be a positive integer"},
    "top_p": {
        "validator": lambda x: is_float(x) and 0 <= float(x) <= 1,
        "msg": "`top_p` must be a float in the interval [0, 1]",
    },
    "temperature": {
        "validator": lambda x: is_float(x) and 0.1 <= float(x) <= 1,
        "msg": "`temperature` must be a float in the interval [0.1, 1]",
    },
    "frequency_penalty": {
        "validator": lambda x: is_float(x) and -2 <= float(x) <= 2,
        "msg": "`frequency_penalty` must be a float in the interval [-2, 2]",
    },
    "presence_penalty": {
        "validator": lambda x: is_float(x) and -2 <= float(x) <= 2,
        "msg": "`presence_penalty` must be a float in the interval [-2, 2]",
    },
}


class RequiredValidator(Validator):
    """Validator to check if a field is required."""

    def __init__(self, field_name: str):
        """Initialize the validator.

        Args:
            field_name: The name of the field to check and print in the error message.
        """
        self.field_name = field_name

    def validate(self, document):
        if not document.text:
            raise ValidationError(message=f"{self.field_name} is required")


class URLValidator(Validator):
    """Validator for URL format."""

    def validate(self, document) -> None:
        text = document.text.strip()
        if not text:
            raise ValidationError(message="URL cannot be empty")
        if not text.startswith(("http://", "https://")):
            raise ValidationError(message="URL must start with http:// or https://")
        if not text.endswith("/"):
            raise ValidationError(message="URL must end with /")
