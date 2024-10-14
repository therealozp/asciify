# Asciify

**Asciify** is a tool for converting images into ASCII art with added edge detection. It uses Go to transform images into ASCII representations, complete with customizable color palettes, edge-based symbol selection, and other stylistic options. This tool can simulate the retro aesthetic of classic text art or experiment with more modern, creative styles.

## Features

- **Image-to-ASCII Conversion**: Converts any image into ASCII characters, where each character represents a block of pixels.
- **Edge Detection Integration**: Utilizes edge detection to replace edge pixels with specific ASCII characters (`|`, `_`, `/`, `\`) to highlight contours.
- **Color Support**: Choose between monochrome, inverted, or full-color outputs.
- **Customizable Characters**: Customize the ASCII characters used for different luminance levels and edges.
- **Gaussian Blur and Bloom Effects**: Simulate additional visual effects like Gaussian blur and bloom for a unique look.
- **CRT and Cyberpunk Effects (Coming Soon)**: Plans for retro CRT filters and neon cyberpunk-inspired aesthetics.

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

## Planned Features

- **Advanced Color Effects**: Retro CRT filter, neon cyberpunk bloom effects.
- **Depth of Field Simulation**: Adds depth perception to the ASCII image for enhanced realism.
- **Customizable ASCII Characters**: Select different characters for a personalized look.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests for new features, improvements, or bug fixes.

## About

Asciify was built as a fun project to explore image processing, Go, and the fascinating world of ASCII art. It combines modern coding techniques with retro aesthetics to create unique and visually appealing images.
