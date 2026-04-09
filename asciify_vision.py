import cv2
import subprocess
from asciify import asciify_from_raw_image
from ultralytics import YOLO

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
                print("Capture approved. Processing...")
                cv2.destroyWindow("Preview (O=Approve, C=Cancel)")

                try:
                    asciify_from_raw_image(
                        best_face_crop,
                        "output_face.txt",
                        max_width_chars=64,
                        max_height_chars=64,
                    )

                    # --- Execute Bash Script ---
                    print(f"Executing bash script: {BASH_SCRIPT_PATH}...")

                    # Uncomment the next line when your bash script is ready:
                    subprocess.run(["/bin/bash", BASH_SCRIPT_PATH], check=True)
                    # print("Bash script executed successfully.")

                except subprocess.CalledProcessError as e:
                    print(f"Bash Script Error: {e}")
                except Exception as e:
                    print(f"An unexpected error occurred: {e}")

                # Go back to webcam stream after processing finishes
                print("Returning to live stream...")
                state = "STREAM"

    # Cleanup
    cap.release()
    cv2.destroyAllWindows()


if __name__ == "__main__":
    main()
