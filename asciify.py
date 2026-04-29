import cv2
import numpy as np
import math
import argparse


def get_edge_direction(angle):
    """
    Returns BGR colors for OpenCV based on quantized angles.
    Note: OpenCV uses BGR, not RGB.
    """
    if angle == 0:
        return (0, 255, 0)  # Green
    elif angle == 45:
        return (0, 255, 255)  # Yellow
    elif angle == 90:
        return (0, 0, 255)  # Red
    elif angle == 135:
        return (255, 0, 0)  # Blue
    return (0, 0, 0)  # Black


def get_angle_heatmap(angle_map):
    """
    Generates a colorized heatmap of the edge angles using NumPy vectorization.
    """
    h, w = angle_map.shape
    img = np.zeros((h, w, 3), dtype=np.uint8)

    # Vectorized color assignment based on angles
    img[angle_map == 0] = (0, 255, 0)
    img[angle_map == 45] = (0, 255, 255)
    img[angle_map == 90] = (0, 0, 255)
    img[angle_map == 135] = (255, 0, 0)

    # cv2.imwrite("angle_heatmap.png", img)
    return img


def optimized_shader_map(angle_map, block_size):
    """
    Generates an ASCII grid based on the dominant angles inside each block.
    """
    h, w = angle_map.shape
    new_h, new_w = h // block_size, w // block_size

    shader_map = [[" " for _ in range(new_w)] for _ in range(new_h)]
    angle_conversions = {0: "-", 45: "/", 90: "|", 135: "\\"}

    for y in range(new_h):
        for x in range(new_w):
            # Extract the block
            y_start, y_end = y * block_size, (y + 1) * block_size
            x_start, x_end = x * block_size, (x + 1) * block_size
            block = angle_map[y_start:y_end, x_start:x_end]

            valid_angles = block[~np.isnan(block)]

            if len(valid_angles) > 0:
                # Find the most common angle and its count
                values, counts = np.unique(valid_angles, return_counts=True)
                max_count_idx = np.argmax(counts)
                max_count = counts[max_count_idx]
                dominant_angle = values[max_count_idx]

                # Check threshold
                if max_count > block_size * 2:
                    shader_map[y][x] = angle_conversions[int(dominant_angle)]

    print("Shader map successfully generated.")
    return shader_map


def compute_shader_map(angle_map, block_size):
    h, w = angle_map.shape
    new_h, new_w = h // block_size, w // block_size
    img = np.zeros((new_h, new_w, 3), dtype=np.uint8)

    for y in range(new_h):
        for x in range(new_w):
            y_start, y_end = y * block_size, (y + 1) * block_size
            x_start, x_end = x * block_size, (x + 1) * block_size
            block = angle_map[y_start:y_end, x_start:x_end]

            valid_angles = block[~np.isnan(block)]

            if len(valid_angles) > 0:
                values, counts = np.unique(valid_angles, return_counts=True)
                max_count_idx = np.argmax(counts)
                max_count = counts[max_count_idx]
                dominant_angle = values[max_count_idx]

                if max_count > block_size * 2:
                    img[y, x] = get_edge_direction(int(dominant_angle))

    # cv2.imwrite("shader_map.png", img)
    return img


def get_sobel_filter(image):
    """
    Calculates the Sobel gradients, magnitudes, and quantized angles.
    """
    if len(image.shape) == 3:
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    else:
        gray = image

    gx = cv2.Sobel(gray, cv2.CV_64F, 1, 0, ksize=3)
    gy = cv2.Sobel(gray, cv2.CV_64F, 0, 1, ksize=3)

    # Compute magnitude and angle in degrees
    magnitude, angle = cv2.cartToPolar(gx, gy, angleInDegrees=True)

    # Normalize magnitude back to uint8 (0-255)
    magnitude = np.clip(magnitude, 0, 255).astype(np.uint8)

    # Wrap angles to 0-180 range (since edges are symmetric)
    angle = angle % 180

    angle_map = np.full(angle.shape, np.nan)

    angle_map[(angle >= 22.5) & (angle < 67.5)] = 45
    angle_map[(angle >= 67.5) & (angle < 112.5)] = 90
    angle_map[(angle >= 112.5) & (angle < 157.5)] = 135
    angle_map[(angle < 22.5) | (angle >= 157.5)] = 0

    # Threshold: wipe out angles where the magnitude is too low
    angle_map[magnitude < 50] = np.nan

    return magnitude, angle_map


def get_luminance_characters(gray_image, reversed=True, gamma=1.0):
    chars = np.array(list(" .:-=+*#%@")[:: -1 if reversed else 1])
    normalized = (gray_image / 255.0) ** gamma  # Apply gamma correction

    indices = np.round(normalized * (len(chars) - 1)).astype(int)
    return chars[indices]


def asciify_to_text(
    image_path,
    output_txt_path,
    scale_factor=4,
    max_width_chars=None,
    max_height_chars=None,
):
    """
    Converts an image to an ASCII text file using Sobel edge maps and luminance.
    """
    img = cv2.imread(image_path)
    if img is None:
        print(f"Error: Could not load image from {image_path}")
        return

    h, w = img.shape[:2]

    new_w = w // scale_factor
    new_h = h // scale_factor

    if max_width_chars is not None and max_height_chars is not None:
        scale_h = int(h / max_height_chars)
        scale_w = int(w / max_width_chars)

        scale_factor = max(scale_h, scale_w)
        new_w = w // scale_factor
        new_h = h // scale_factor

    downscaled = cv2.resize(img, (new_w, new_h), interpolation=cv2.INTER_AREA)
    gray_downscaled = cv2.cvtColor(downscaled, cv2.COLOR_BGR2GRAY)

    _, angle_map = get_sobel_filter(img)
    edge_map = optimized_shader_map(angle_map, scale_factor)

    lum_chars = get_luminance_characters(gray_downscaled)

    # 4. Merge the edge map and luminance map, then write to file
    with open(output_txt_path, "w", encoding="utf-8") as f:
        f.write("\eA\x05\e2\r\n")
        for y in range(new_h):
            row_chars = []
            for x in range(new_w):
                if edge_map[y][x] != " ":
                    row_chars.append(edge_map[y][x])
                else:
                    row_chars.append(lum_chars[y, x])

            f.write("".join(row_chars) + "\r\n")

    print(f"Successfully generated ASCII text file at: {output_txt_path}")


def asciify_from_raw_image(
    raw_image,
    output_txt_path,
    scale_factor=4,
    max_width_chars=None,
    max_height_chars=None,
    gamma=1.0,
):
    """
    Converts an image to an ASCII text file using Sobel edge maps and luminance.
    """
    img = raw_image
    if img is None:
        print("Error: Could not load image.")
        return

    h, w = img.shape[:2]

    new_w = w // scale_factor
    new_h = h // scale_factor

    if max_width_chars is not None and max_height_chars is not None:
        scale_h = int(h / max_height_chars)
        scale_w = int(w / max_width_chars)

        scale_factor = max(scale_h, scale_w)
        new_w = w // scale_factor
        new_h = h // scale_factor

    downscaled = cv2.resize(img, (new_w, new_h), interpolation=cv2.INTER_AREA)
    gray_downscaled = cv2.cvtColor(downscaled, cv2.COLOR_BGR2GRAY)

    _, angle_map = get_sobel_filter(img)
    edge_map = optimized_shader_map(angle_map, scale_factor)

    lum_chars = get_luminance_characters(gray_downscaled, gamma=gamma)
    lines = []
    # 4. Merge the edge map and luminance map, then write to file
    if output_txt_path is None:
        for y in range(new_h):
            row_chars = []
            for x in range(new_w):
                if edge_map[y][x] != " ":
                    row_chars.append(edge_map[y][x])
                else:
                    row_chars.append(lum_chars[y, x])
            lines.append("".join(row_chars))
        return lines

    with open(output_txt_path, "w", encoding="utf-8") as f:
        f.write("\eA\x05\e2\r\n")
        for y in range(new_h):
            row_chars = []
            for x in range(new_w):
                if edge_map[y][x] != " ":
                    row_chars.append(edge_map[y][x])
                else:
                    row_chars.append(lum_chars[y, x])
            lines.append("".join(row_chars))
            f.write("".join(row_chars) + "\r\n")

    print(f"Successfully generated ASCII text file at: {output_txt_path}")
    return lines


if __name__ == "__main__":
    # Set up the argument parser
    parser = argparse.ArgumentParser(
        description="Convert an image to ASCII text using Sobel edge detection."
    )

    parser.add_argument(
        "input", help="Path to the input image file (e.g., input_image.jpg)"
    )

    parser.add_argument(
        "-o",
        "--output",
        default="output_ascii.txt",
        help="Path to the output text file (default: output_ascii.txt)",
    )
    parser.add_argument(
        "-s",
        "--scale",
        type=int,
        default=4,
        help="Scale factor for downsampling (default: 4)",
    )
    parser.add_argument(
        "-W",
        "--width",
        type=int,
        default=50,
        help="Maximum width in characters (default: 50)",
    )
    parser.add_argument(
        "-H",
        "--height",
        type=int,
        default=28,
        help="Maximum height in characters (default: 28)",
    )

    args = parser.parse_args()

    asciify_to_text(
        args.input,
        args.output,
        scale_factor=args.scale,
        max_width_chars=args.width,
        max_height_chars=args.height,
    )
