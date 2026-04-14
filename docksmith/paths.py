# docksmith/paths.py
# Centralised path constants — import from here instead of repeating expanduser() everywhere.

import os

DOCKSMITH_DIR = os.path.expanduser("~/.docksmith")
IMAGES_DIR    = os.path.join(DOCKSMITH_DIR, "images")
LAYERS_DIR    = os.path.join(DOCKSMITH_DIR, "layers")
CACHE_DIR     = os.path.join(DOCKSMITH_DIR, "cache")
CACHE_INDEX   = os.path.join(CACHE_DIR, "index.json")
BUILD_LOG     = os.path.join(DOCKSMITH_DIR, "build_log.json")


def ensure_all():
    for d in [IMAGES_DIR, LAYERS_DIR, CACHE_DIR]:
        os.makedirs(d, exist_ok=True)


def docksmith_dir() -> str:
    return DOCKSMITH_DIR


def images_dir() -> str:
    return IMAGES_DIR


def layers_dir() -> str:
    return LAYERS_DIR


def cache_dir() -> str:
    return CACHE_DIR
