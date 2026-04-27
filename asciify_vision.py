import cv2
import subprocess
from asciify import asciify_from_raw_image
from ultralytics import YOLO
import numpy as np

# --- CONFIGURATION ---
MODEL_PATH = "yolov8n-face.pt"  # Replace with your local YOLO face model
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

                font = cv2.FONT_HERSHEY_PLAIN
                font_scale = 0.6
                thickness = 1
                char_w, char_h = 6, 10

                gamma = 1.0
                GAMMA_STEP = 0.1

                def render_ascii_canvas(face_crop, gamma):
                    lines = asciify_from_raw_image(
                        face_crop,
                        output_txt_path=None,
                        max_width_chars=64,
                        max_height_chars=64,
                        gamma=gamma,
                    )
                    if not lines:
                        return None, lines

                    canvas_h = len(lines) * char_h + 10
                    canvas_w = max(len(l) for l in lines) * char_w + 10
                    canvas = np.zeros((canvas_h, canvas_w, 3), dtype=np.uint8)

                    for i, line in enumerate(lines):
                        cv2.putText(
                            canvas,
                            line,
                            (4, (i + 1) * char_h),
                            font,
                            font_scale,
                            (200, 200, 200),
                            thickness,
                        )

                    # HUD: show current gamma
                    label = f"gamma={gamma:.2f}  (+/- to adjust, O=Save, C=Cancel)"
                    cv2.putText(
                        canvas,
                        label,
                        (4, canvas_h - 2),
                        font,
                        font_scale,
                        (80, 200, 80),
                        thickness,
                    )

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
