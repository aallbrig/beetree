import unittest

import validate_findings as v

GOOD_CONFIRMED = {
    "id": "AUTH-1", "unit": "auth", "title": "Harden token check", "verdict": "confirmed",
    "severity": "low", "trace": ["api/auth.go:10"], "observed": "test fails", "fix": "harden; add test",
}


class ValidateFindings(unittest.TestCase):
    def test_valid_all_verdicts(self):
        data = [
            GOOD_CONFIRMED,
            {"id": "N-1", "unit": "auth", "title": "t", "verdict": "needs_validation", "unresolved_fact": "is X set"},
            {"id": "R-1", "unit": "auth", "title": "t", "verdict": "rejected", "reason": "guarded"},
        ]
        self.assertEqual(v.check(data), [])

    def test_not_a_list(self):
        self.assertTrue(v.check({"a": 1}))

    def test_bad_verdict(self):
        self.assertTrue(v.check([{**GOOD_CONFIRMED, "verdict": "maybe"}]))

    def test_confirmed_needs_trace_and_severity(self):
        bad = {k: x for k, x in GOOD_CONFIRMED.items() if k not in ("trace", "severity")}
        self.assertEqual(len(v.check([bad])), 4)

    def test_trace_must_have_line(self):
        self.assertTrue(v.check([{**GOOD_CONFIRMED, "trace": ["api/auth.go"]}]))

    def test_severity_forbidden_on_needs_validation(self):
        f = {"id": "N", "unit": "u", "title": "t", "verdict": "needs_validation", "unresolved_fact": "x", "severity": "low"}
        self.assertTrue(v.check([f]))

    def test_duplicate_id(self):
        self.assertTrue(v.check([GOOD_CONFIRMED, GOOD_CONFIRMED]))

    def test_leaks_rejected(self):
        for text in ("acct 123456789012", "see /home/someone/x", "mail bob@corp.io"):
            self.assertTrue(v.check([{**GOOD_CONFIRMED, "observed": text}]), text)
        self.assertEqual(v.check([{**GOOD_CONFIRMED, "observed": "you@example.com"}]), [])


if __name__ == "__main__":
    unittest.main()
