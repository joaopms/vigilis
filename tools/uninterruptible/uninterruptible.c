#include<stdio.h>
#include<string.h>
#include<unistd.h>

int main(int argc, char *argv[]) {
    // Skip if the version flag is given, so Vigilis doesn't hang when checking ffmpeg's version
    for (int i = 0; i < argc; i++) {
        if (!strcmp("-version", argv[i])) { // returns zero when equal
            printf("ffmpeg version 0.0-uninterruptible\n");
            printf("Got version flag, exiting\n");
            return 0;
        }
    }

    printf("Created an uninterruptible process with PID %d\n", getpid());
    printf("Waiting for a signal to exit\n");
    vfork();
    pause();
    return 0;
}