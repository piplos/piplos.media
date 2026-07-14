#!/usr/bin/env python3
"""Remove a flat magenta (#FF00FF) chroma-key background from generated illustrations.

Usage:
    python3 scripts/chroma_key.py input.png output.png

The output is an RGBA PNG cropped to content with 8px padding.
Used for Piplos Media brand illustrations (see docs/illustrations.md).
"""

import sys

import numpy as np
from PIL import Image


def chroma_key(input_path: str, output_path: str) -> None:
    img = Image.open(input_path).convert('RGB')
    a = np.asarray(img).astype(np.float32)
    r, g, b = a[..., 0], a[..., 1], a[..., 2]

    # Distance from pure magenta: high R, low G, high B.
    dist = np.sqrt((r - 255) ** 2 + g ** 2 + (b - 255) ** 2)
    # <60 fully transparent, >180 fully opaque, linear ramp in between.
    alpha = np.clip((dist - 60) / 120, 0, 1)

    # Defringe: remove magenta cast everywhere (also opaque shadow pixels).
    # The brand palette contains no magenta, so min(r,b) > g is always chroma bleed.
    mag_cast = np.clip(np.minimum(r, b) - g, 0, 255)
    r2 = np.clip(r - mag_cast, 0, 255)
    b2 = np.clip(b - mag_cast, 0, 255)

    out = np.stack([r2, g, b2, alpha * 255], axis=-1).astype(np.uint8)
    res = Image.fromarray(out, 'RGBA')

    bbox = res.getbbox()
    pad = 8
    bbox = (
        max(0, bbox[0] - pad),
        max(0, bbox[1] - pad),
        min(res.width, bbox[2] + pad),
        min(res.height, bbox[3] + pad),
    )
    res = res.crop(bbox)
    res.save(output_path)
    print(f'saved {output_path} {res.size} {res.mode}')


if __name__ == '__main__':
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    chroma_key(sys.argv[1], sys.argv[2])
