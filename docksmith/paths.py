import os


def _is_writable_dir(path: str) -> bool:
    """Best-effort check whether a directory can be created/written."""
    try:
        os.makedirs(path, exist_ok=True)
        probe = os.path.join(path, ".docksmith_write_probe")
        with open(probe, "w", encoding="utf-8") as f:
            f.write("ok")
        os.remove(probe)
        return True
    except OSError:
        return False


def get_docksmith_dir() -> str:
    """
    Returns Docksmith state directory with safe fallback.

    Priority:
    1) $DOCKSMITH_DIR
    2) ~/.docksmith
    3) ./.docksmith (workspace fallback for restricted environments)
    """
    env_dir = os.environ.get("DOCKSMITH_DIR")
    if env_dir:
        return os.path.abspath(os.path.expanduser(env_dir))

    home_default = os.path.abspath(os.path.expanduser("~/.docksmith"))
    if _is_writable_dir(home_default):
        return home_default

    return os.path.abspath(os.path.join(os.getcwd(), ".docksmith"))


def images_dir() -> str:
    return os.path.join(get_docksmith_dir(), "images")


def layers_dir() -> str:
    return os.path.join(get_docksmith_dir(), "layers")


def cache_dir() -> str:
    return os.path.join(get_docksmith_dir(), "cache")
