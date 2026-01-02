"""RTMX - Requirements Traceability Matrix toolkit for GenAI-driven development.

This package provides tools for managing requirements traceability in software projects,
with special focus on compliance frameworks (CMMC, FedRAMP) and GenAI integration.

Example:
    >>> from rtmx import RTMDatabase, Status
    >>> db = RTMDatabase.load("docs/rtm_database.csv")
    >>> incomplete = db.filter(status=Status.MISSING)
    >>> cycles = db.find_cycles()
"""

from rtmx.config import RTMXConfig, load_config
from rtmx.models import (
    Priority,
    Requirement,
    RequirementNotFoundError,
    RTMDatabase,
    RTMError,
    RTMValidationError,
    Status,
)

__version__ = "0.0.2"
__all__ = [
    "RTMDatabase",
    "Requirement",
    "Status",
    "Priority",
    "RTMError",
    "RequirementNotFoundError",
    "RTMValidationError",
    "RTMXConfig",
    "load_config",
    "__version__",
]
