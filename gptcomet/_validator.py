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
}
