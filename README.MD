# Project 3 - N-body Simulation
by Nesta Lenhert-Scholer

# Sources
I would like to start by thanking the following sources for help during this project:

[Coding Train Quadtree Youtube](https://www.youtube.com/watch?v=OJxEcs0w_kE)  
[Princeton Barnes-Hut Assignment](https://www.cs.princeton.edu/courses/archive/fall03/cs126/assignments/barnes-hut.html)

The only external library used is [raylib-go](https://github.com/gen2brain/raylib-go). It is primarily
used for the GUI interface as well as mathematical vector objects and operations.

# How to Run It
The following is the usage statement of the program:  
```
Usage: ./sim [-w | -i=INTEGER] <X> <Y> <thread_count>
            -w = Run this program in GUI mode.
            -i = Number of updates to run. Must be greater than 0. (Note only have -w or -i, not both.
            <X> = The width of the window. Positive Integer.
            <Y> = The height of the window. Positive Integer.
            <thread_count> = Number of maximum threads to use. Set to 0 to run in sequential mode.
```

The first command specifies whether to run it in GUI mode or sequential mode.
Type `-w` for GUI mode, or type `-i=INT` to run in console mode for `INT` number of
time-steps. The next two arguments specify the dimensions that the objects are in, 
as well as the screen size if in GUI mode. The final argument specifies how many threads
to run. If in console mode, setting this to `0` will run the program in sequential mode.
For GUI mode, this number needs to be greater than `0` since it supports both parallel
and sequential mode. 

In GUI mode, press key `SPACE` to change between Sequential and Parallel modes. Press key `B`
to show the internal Barnes-Hut Tree.

Data can be generated using [generate.go](proj3/generate.go). Here, the only arguments
are the number of objects to create, and the dimensions. Make sure that the dimensions
used to generate the data are the same as the dimensions when running the program.

Three shell scripts are provided to run the GUI (*[runGUI.sh](proj3/runGUI.sh)*), sequential
(*[runConsole_seq.sh](proj3/runConsole_seq.sh)*), and parallel
(*[runConsole_par.sh](proj3/runConsole_par.sh)*) versions on some test
data.

Thank you for the great class!

