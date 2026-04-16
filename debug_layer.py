#!/usr/bin/env python3
import tarfile
import os

# Find the latest layer (should be the COPY layer)
layers_dir = os.path.expanduser('~/.docksmith/layers')
layers = sorted(os.listdir(layers_dir))

print(f"Total layers: {len(layers)}\n")

for layer_file in layers:
    print(f"Layer: {layer_file}")
    layer_path = os.path.join(layers_dir, layer_file)
    with tarfile.open(layer_path, 'r:') as tar:
        members = tar.getmembers()
        print(f"  Files: {len(members)}")
        for member in members[:10]:  # Show first 10
            print(f"    {member.name} (type: {member.type})")
        if len(members) > 10:
            print(f"    ... and {len(members) - 10} more")
    print()
