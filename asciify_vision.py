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
                print("Capture approved. Processing ASCII preview...")
                cv2.destroyWindow("Preview (O=Approve, C=Cancel)")

                # Generate ASCII but don't write file yet
                lines = asciify_from_raw_image(
                    best_face_crop,
                    output_txt_path=None,  # skip writing
                    max_width_chars=64,
                    max_height_chars=64,
                )

                if lines:
                    # Render ASCII into an OpenCV image for preview
                    font = cv2.FONT_HERSHEY_PLAIN
                    font_scale = 0.6
                    thickness = 1
                    char_w, char_h = 6, 10  # approx px per char at this scale

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

                    cv2.imshow("ASCII Preview (O=Save, C=Cancel)", canvas)

                    while True:
                        k = cv2.waitKey(0) & 0xFF
                        if k == ord("o"):
                            cv2.destroyWindow("ASCII Preview (O=Save, C=Cancel)")
                            # Now write the file
                            asciify_from_raw_image(
                                best_face_crop,
                                "output_face.txt",
                                max_width_chars=64,
                                max_height_chars=64,
                            )
                            try:
                                subprocess.run(
                                    ["/bin/bash", BASH_SCRIPT_PATH], check=True
                                )
                            except subprocess.CalledProcessError as e:
                                print(f"Bash Script Error: {e}")
                            break
                        elif k == ord("c"):
                            cv2.destroyWindow("ASCII Preview (O=Save, C=Cancel)")
                            print("ASCII discarded. Returning to stream...")
                            break

                state = "STREAM"

    # Cleanup
    cap.release()
    cv2.destroyAllWindows()


if __name__ == "__main__":
    main()
