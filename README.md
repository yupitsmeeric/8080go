QUICK RESTART:
How to compile:
run `go build`

NOTES:
clock speed is 2MHz, delay is 5e-7s = 5e-4ms

TODO:
- run and compare to a reference implementation to make sure that the stuff works
reference:
https://8080.cakers.io/

- ebitengine:
    - make some kind of loop that connects the screen to the cpu cycles - 1 screen cycle runs 1 second worth of cpu cycles
    - or maybe modify the run function to specify number of cycles to run
https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2

- Connect the screen to the system
- Make a function that steps forward the screen with the CPU cycles:
https://www.reddit.com/r/EmuDev/comments/ksbvgx/how_to_set_clock_speed_in_c/

RESOURCES:
https://web.archive.org/web/20240118230906/http://www.emulator101.com/full-8080-emulation.html
https://web.archive.org/web/20240118230900/http://www.emulator101.com/reference/8080-by-opcode.html
reference 8080:
https://github.com/hlboehm/i8080-emulator/blob/main/src/emulator/processor.c
https://tobiasvl.github.io/optable/intel-8080/
for stuff outside the cpu
https://computerarcheology.com/Arcade/SpaceInvaders/Hardware.html
https://www.reddit.com/r/EmuDev/comments/ksbvgx/how_to_set_clock_speed_in_c/

