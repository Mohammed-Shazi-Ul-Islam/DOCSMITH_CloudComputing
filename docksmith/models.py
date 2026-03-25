# docksmith/models.py
# ============================================================
#  PIYUSH — File 1: Shared Data Models
#  Write this FIRST. Share with team immediately.
#  Everyone imports ImageManifest, LayerEntry, ImageConfig
# ============================================================

from dataclasses import dataclass, field
from typing import List


@dataclass
class ImageConfig:
    """
    Stores the runtime configuration of an image.
    Env:       list of "KEY=value" strings  e.g. ["APP=myapp", "PORT=8080"]
    Cmd:       list of strings              e.g. ["python", "main.py"]
    WorkingDir: string path                 e.g. "/app"
    """
    Env:        List[str] = field(default_factory=list)
    Cmd:        List[str] = field(default_factory=list)
    WorkingDir: str       = ""


@dataclass
class LayerEntry:
    """
    Represents one layer (delta tar) of an image.
    digest:    content-addressed ID  e.g. "sha256:abc123..."
    size:      byte size of the tar file
    createdBy: the instruction that produced this layer  e.g. "COPY . /app"
    """
    digest:    str
    size:      int
    createdBy: str = ""


@dataclass
class ImageManifest:
    """
    The full record for a stored image.
    Saved to ~/.docksmith/images/<name>/<tag>.json
    
    name:    image name   e.g. "myapp"
    tag:     image tag    e.g. "latest"
    digest:  SHA-256 of the manifest JSON (computed by save_manifest)
    created: ISO-8601 timestamp string
    config:  ImageConfig object
    layers:  list of LayerEntry objects
    """
    name:    str
    tag:     str
    digest:  str
    created: str
    config:  ImageConfig
    layers:  List[LayerEntry] = field(default_factory=list)