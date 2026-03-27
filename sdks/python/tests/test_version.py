from muga import __version__


def test_version_is_string() -> None:
    assert isinstance(__version__, str)


def test_version_format() -> None:
    parts = __version__.split(".")
    assert len(parts) == 3
    for part in parts:
        assert part.isdigit()
