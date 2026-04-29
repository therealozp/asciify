import cv2
import subprocess
from asciify import asciify_from_raw_image
from ultralytics import YOLO
from PIL import Image, ImageDraw, ImageFont
import numpy as np

# --- CONFIGURATION ---
MODEL_PATH = "yolov12n-face.pt"  # Replace with your local YOLO face model
BASH_SCRIPT_PATH = "./run_after_api.sh"


def main():
    # 1. Load the Ultralytics YOLO model
    print("Loading YOLO model...")
    model = YOLO(MODEL_PATH)

    # Initialize webcam
    cap = cv2.VideoCapture(0)
    if not cap.isOpened():
        print("Error: Could not open webcam.")
        return

    state = "STREAM"  # States: 'STREAM' or 'PREVIEW'
    best_face_crop = None

    print("Webcam started. Press SPACE to capture, Q to quit.")

    while True:
        if state == "STREAM":
            ret, frame = cap.read()
            if not ret:
                break

            # Keep a clean copy of the frame to crop from (without drawn boxes)
            clean_frame = frame.copy()

            # Run YOLO inference
            results = model(frame, stream=True, verbose=False)

            best_box = None
            max_conf = 0.0

            # Find the single face with the highest confidence
            for result in results:
                for box in result.boxes:
                    conf = float(box.conf[0])
                    if conf > max_conf:
                        max_conf = conf
                        best_box = box

            # If a face is found, draw the bounding box and prep the crop
            if best_box is not None:
                x1, y1, x2, y2 = map(int, best_box.xyxy[0])
                expansion = 0.2
                dist_x = int((x2 - x1) * expansion)
                dist_y = int((y2 - y1) * expansion)

                x1 = max(0, x1 - dist_x // 2)
                y1 = max(0, y1 - dist_y // 2)
                x2 = min(frame.shape[1], x2 + dist_x // 2)
                y2 = min(frame.shape[0], y2 + dist_y // 2)

                # Draw the bounding box on the display frame
                cv2.rectangle(frame, (x1, y1), (x2, y2), (0, 255, 0), 2)
                cv2.putText(
                    frame,
                    f"Conf: {max_conf:.2f}",
                    (x1, y1 - 10),
                    cv2.FONT_HERSHEY_SIMPLEX,
                    0.5,
                    (0, 255, 0),
                    2,
                )

                # Extract the crop from the clean frame
                # Optional: You can add padding to x1, y1, x2, y2 here if the crop is too tight
                best_face_crop = clean_frame[y1:y2, x1:x2]

            # Show the live stream
            cv2.imshow("Live Stream", frame)

            # Wait for key press
            key = cv2.waitKey(1) & 0xFF

            if key == ord("q"):  # Quit
                break
            elif key == ord(" "):  # SPACE pressed to capture
                if best_face_crop is not None and best_face_crop.size > 0:
                    state = "PREVIEW"
                else:
                    print("No face detected to capture!")

        elif state == "PREVIEW":
            # 2. Freeze frame and show preview
            cv2.imshow("Preview (O=Approve, C=Cancel)", best_face_crop)

            key = cv2.waitKey(1) & 0xFF

            # 3. Handle Cancel ('C')
            if key == ord("c"):
                print("Capture cancelled. Returning to stream...")
                cv2.destroyWindow("Preview (O=Approve, C=Cancel)")
                state = "STREAM"

            # 3. Handle Approve ('O')
            elif key == ord("o"):
                print("Capture approved. Generating ASCII preview...")
                cv2.destroyWindow("Preview (O=Approve, C=Cancel)")

                gamma = 1.0
                GAMMA_STEP = 0.1

                def render_ascii_canvas(
                    face_crop,
                    gamma,
                    font_path="assets/cpc464",
                    font_size=12,
                ):
                    lines = asciify_from_raw_image(
                        face_crop,
                        output_txt_path=None,
                        max_width_chars=64,
                        max_height_chars=64,
                        gamma=gamma,
                    )
                    if not lines:
                        return None, lines

                    try:
                        font = ImageFont.truetype(font_path, font_size)
                    except IOError:
                        print(
                            f"Error: Could not load font at {font_path}. Falling back to default."
                        )
                        font = ImageFont.load_default()

                    left, top, right, bottom = font.getbbox("A")
                    char_w = right - left
                    char_h = bottom - top

                    canvas_h = len(lines) * char_h + 40
                    canvas_w = max(len(l) for l in lines) * char_w + 20

                    # 3. Create a blank PIL Image (RGB format)
                    pil_canvas = Image.new("RGB", (canvas_w, canvas_h), color=(0, 0, 0))
                    draw = ImageDraw.Draw(pil_canvas)

                    for i, line in enumerate(lines):
                        y_position = (i * char_h) + 10
                        draw.text(
                            (4, y_position),
                            line,
                            font=font,
                            fill=(200, 200, 200),  # RGB Color
                        )

                    label = f"gamma={gamma:.2f}  (+/- to adjust, O=Save, C=Cancel)"
                    draw.text(
                        (4, canvas_h - char_h - 10),
                        label,
                        font=font,
                        fill=(80, 200, 80),  # RGB Green
                    )

                    canvas = np.array(pil_canvas)
                    canvas = cv2.cvtColor(canvas, cv2.COLOR_RGB2BGR)

                    return canvas, lines

                canvas, lines = render_ascii_canvas(best_face_crop, gamma)

                if canvas is not None:
                    cv2.imshow("ASCII Preview", canvas)

                    while True:
                        k = cv2.waitKey(0) & 0xFF

                        if k == ord("o"):
                            cv2.destroyWindow("ASCII Preview")
                            # Write file with current gamma
                            asciify_from_raw_image(
                                best_face_crop,
                                "output_face.txt",
                                max_width_chars=64,
                                max_height_chars=64,
                                gamma=gamma,
                            )
                            try:
                                subprocess.run(
                                    ["/bin/bash", BASH_SCRIPT_PATH], check=True
                                )
                            except subprocess.CalledProcessError as e:
                                print(f"Bash Script Error: {e}")
                            break

                        elif k == ord("c"):
                            cv2.destroyWindow("ASCII Preview")
                            print("ASCII discarded. Returning to stream...")
                            break

                        elif k == ord("+") or k == ord("="):  # = is unshifted +
                            gamma = round(min(gamma + GAMMA_STEP, 5.0), 2)
                            canvas, lines = render_ascii_canvas(best_face_crop, gamma)
                            cv2.imshow("ASCII Preview", canvas)

                        elif k == ord("-"):
                            gamma = round(max(gamma - GAMMA_STEP, 0.1), 2)
                            canvas, lines = render_ascii_canvas(best_face_crop, gamma)
                            cv2.imshow("ASCII Preview", canvas)

                state = "STREAM"

    # Cleanup
    cap.release()
    cv2.destroyAllWindows()


if __name__ == "__main__":
    main()
