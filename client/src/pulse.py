import time
import os

os.system("pkill pulseaudio")
time.sleep(3)
os.system("pulseaudio & disown")
print("Rebooted Pulse")
