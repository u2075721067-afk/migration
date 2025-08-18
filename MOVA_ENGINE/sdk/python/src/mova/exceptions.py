"""MOVA SDK exceptions."""


class MOVAError(Exception):
    """Base exception for MOVA SDK."""

    pass


class MOVAAPIError(MOVAError):
    """Exception raised when API returns an error."""

    pass


class MOVAConnectionError(MOVAError):
    """Exception raised when connection to API fails."""

    pass


class MOVATimeoutError(MOVAError):
    """Exception raised when request times out."""

    pass


class MOVAValidationError(MOVAAPIError):
    """Exception raised when envelope validation fails."""

    pass
