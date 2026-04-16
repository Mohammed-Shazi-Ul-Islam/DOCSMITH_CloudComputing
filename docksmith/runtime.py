import os
import re
import subprocess
import sys
import tempfile
from typing import Dict, Iterable, List, Optional

from docksmith.layers import extract_layer


def _parse_env_pairs(pairs: Optional[Iterable[str]]) -> Dict[str, str]:
    env: Dict[str, str] = {}
    for item in pairs or []:
        if "=" in item:
            key, value = item.split("=", 1)
            env[key] = value
    return env


def _resolve_workdir(rootfs: str, workdir: str) -> str:
    """Resolve and create the working directory inside the rootfs."""
    rel = (workdir or "/").lstrip("/")
    host_path = os.path.join(rootfs, rel)
    os.makedirs(host_path, exist_ok=True)
    return host_path


def _can_attempt_namespace_isolation() -> bool:
    return (
        os.name == "posix"
        and hasattr(os, "chroot")
        and hasattr(os, "unshare")
        and hasattr(os, "CLONE_NEWNS")
        and hasattr(os, "CLONE_NEWPID")
        and hasattr(os, "CLONE_NEWUTS")
        and hasattr(os, "geteuid")
        and os.geteuid() == 0
    )


def _namespace_preexec(rootfs: str, workdir: str):
    def _fn():
        flags = os.CLONE_NEWNS | os.CLONE_NEWPID | os.CLONE_NEWUTS
        os.unshare(flags)
        os.chroot(rootfs)
        os.chdir(workdir or "/")
        os.environ["PS1"] = "(docksmith) "

    return _fn


def isolate_and_run(
    rootfs: str,
    command: List[str],
    env: Optional[Dict[str, str]] = None,
    workdir: str = "/",
) -> int:
    """
    Executes command against assembled rootfs.

    On Linux as root, attempts namespace + chroot isolation.
    Otherwise, executes with cwd rooted under extracted rootfs (safe fallback).
    """
    if not command:
        raise RuntimeError("[RUNTIME ERROR] Empty command")

    merged_env = os.environ.copy()
    merged_env.update(env or {})

    host_workdir = _resolve_workdir(rootfs, workdir)
    
    # Check if we can use namespace isolation (requires root)
    if _can_attempt_namespace_isolation():
        # Running as root - use full isolation with chroot
        preexec_fn = _namespace_preexec(rootfs, workdir or "/")
        cwd = None
    else:
        # Not running as root - adjust command for non-chroot execution
        adjusted_command = []
        for arg in command:
            if arg in ["/bin/sh", "/bin/bash"]:
                # Use host shell
                adjusted_command.append("sh")
            elif arg.startswith('/') and not arg.startswith('/bin/') and not arg.startswith('/usr/bin/'):
                # Convert absolute paths (except binaries) to rootfs paths
                rootfs_path = os.path.join(rootfs, arg.lstrip('/'))
                adjusted_command.append(rootfs_path)
            else:
                adjusted_command.append(arg)
        
        command = adjusted_command
        preexec_fn = None
        cwd = host_workdir

    try:
        proc = subprocess.Popen(
            command,
            cwd=cwd,
            env=merged_env,
            preexec_fn=preexec_fn,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        stdout, stderr = proc.communicate()
        
        # Print output for debugging
        if stdout:
            print(stdout.decode('utf-8', errors='ignore'), end='')
        if stderr:
            print(stderr.decode('utf-8', errors='ignore'), end='', file=sys.stderr)
        
        return proc.returncode
    except FileNotFoundError as exc:
        raise RuntimeError(f"[RUNTIME ERROR] Command not found: {command[0]}") from exc
    except PermissionError as exc:
        raise RuntimeError(
            "[RUNTIME ERROR] Permission denied while starting isolated process."
        ) from exc


def run_image(manifest, command_override: Optional[List[str]] = None, env_overrides: Optional[Dict[str, str]] = None) -> int:
    """
    Runtime entrypoint used by CLI:
    1) assemble rootfs from layers
    2) resolve command and env
    3) run via shared isolation primitive
    """
    command = command_override or list(manifest.config.Cmd or [])
    if not command:
        raise RuntimeError(
            "[RUNTIME ERROR] No command provided and image has no CMD."
        )

    env = _parse_env_pairs(manifest.config.Env)
    env.update(env_overrides or {})
    workdir = manifest.config.WorkingDir or "/"

    with tempfile.TemporaryDirectory(prefix="docksmith_run_") as rootfs:
        for layer in manifest.layers:
            extract_layer(layer.digest, rootfs)
        return isolate_and_run(
            rootfs=rootfs,
            command=command,
            env=env,
            workdir=workdir,
        )
