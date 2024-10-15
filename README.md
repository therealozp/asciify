# asciify

**asciify** is a tool for converting images into ASCII art with added edge detection. It uses Go to transform images into ASCII representations, complete with customizable color palettes, edge-based symbol selection, and other stylistic options.

This project was inspired Acerola's excellent YouTube video on making shaders: [I Tried Turning Games into Text](https://www.youtube.com/watch?v=gg40RWiaHRY&t=917s). Massive thanks to Acerola for the inspiration, and hurricane Milton for finally spurring me to do this.

## Features

[x] **Luminance based**: Converts any image into ASCII characters, strictly using the luminance value of a block of pixels.
[x] **Color Support**: Choose between monochrome, inverted, or full-color outputs.
[x] **Edge Detection**: Utilizes Sobel filter to perform edge detection to replace edge pixels with specific ASCII characters (`|`, `_`, `/`, `\`) to highlight contours.
[x] **Difference-of-gaussians preprocessing**: Used as a preprocessing step to filter out extra variations and contours, leaving only the most pronounced one for the edge detection algorithm
[x] **Effects suite**: Loaded with bloom, color burn, effects for more pronounced color processing.

## Planned

[ ] **CLI**: to have stuff be generated dynamically instead of having to manually adjust parameters.
[ ] **Customizable Characters**: Customize the ASCII characters used for different luminance levels and edges.
[ ] **CRT Effect**: Plans for retro CRT filters and neon cyberpunk-inspired aesthetics.

## Installation

To use Asciify, you'll need to have Go installed. Clone this repository, then build and run the tool:

```bash
git clone https://github.com/yourusername/asciify.git
cd asciify
go build
./asciify
```

## Usage

You can convert an image to ASCII art using the following command:

```bash
./asciify -input /path/to/image.png -output /path/to/output.png -width 100 -height 50 -scaleFactor 2 -monochrome true -inverted false -font /path/to/font.ttf
```

### Command-Line Options

- `-input`: Path to the input image.
- `-output`: Path where the output ASCII art will be saved.
- `-width`: Width of the ASCII art in characters.
- `-height`: Height of the ASCII art in characters.
- `-scaleFactor`: Factor by which to scale the output.
- `-monochrome`: If true, output is monochrome. If false, retains original colors.
- `-inverted`: Invert the brightness of characters in monochrome mode.
- `-font`: Path to a custom font file (TTF) to use for rendering ASCII characters.

## Example

```bash
./asciify -input example.jpg -output ascii_art.png -width 80 -height 40 -scaleFactor 2 -monochrome false -inverted false -font ./fonts/monospace.ttf
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests for new features, improvements, or bug fixes.
