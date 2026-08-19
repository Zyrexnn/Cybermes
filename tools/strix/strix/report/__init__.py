"""Report/finding helpers."""

from strix.report.dedupe import check_duplicate
from strix.report.state import ReportState, get_global_report_state, set_global_report_state


__all__ = [
    "ReportState",
    "check_duplicate",
    "get_global_report_state",
    "set_global_report_state",
]
