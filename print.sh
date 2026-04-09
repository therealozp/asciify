cat output_ascii.txt | xargs -0 echo -en | sudo tee /dev/usb/lp0
