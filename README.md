# PBM_ROTATE

Rotates any PBM images with P1 magic number to any given degrees.

## Usage

```bash
Usage: pbm_rotate [options...]
Examples:
  # 
        pbm_rotate --file foo.pbm --angle 90
Options:
 -f --file      Path to a PBM file (only Magic Number P1 is supported).
 -a --angle     Angle of rotation (in degrees 90,-90 etc.)
```

## Known Limitations

Since we used Rotation Matrix to rotate the image, it is not a lossless algorithm. So for small images some degrees can cause loss of information, but for larger ones it is more acceptable.

<https://en.wikipedia.org/wiki/Rotation_matrix>
