#!/bin/bash
# AITriage - VOGUE PROD LUXE DOCKER BUILDER

# Hide cursor and ensure cleanup
trap 'tput cnorm; echo ""; exit' INT TERM EXIT
tput civis

# Silent Luxury Colors
DIM="\033[38;5;238m"
GRAY="\033[38;5;245m"
WHITE="\033[38;5;255m"
ACCENT="\033[38;5;44m"
RESET="\033[0m"
BOLD="\033[1m"
GREEN="\033[38;5;108m"
RED="\033[38;5;131m"

STEPS=(
    "Allocating build context"
    "Resolving upstream dependencies"
    "Compiling AITriage core binary"
    "Installing security scanners (semgrep, trivy, gitleaks, bandit)"
    "Applying security hardenings"
    "Materializing final container"
)
TOTAL_STEPS=${#STEPS[@]}

TMP_LOG=$(mktemp)
docker build -q -t aitriage:latest . > "$TMP_LOG" 2>&1 &
PID=$!

SPINNER=("⣾" "⣽" "⣻" "⢿" "⡿" "⣟" "⣯" "⣷")
SPINNER_LEN=${#SPINNER[@]}

tick=0
current_step=0
frame_idx=0

# Save current cursor position to return to it
tput sc

while kill -0 $PID 2>/dev/null; do
    # Restore position
    tput rc
    
    # Draw header
    printf "  ${ACCENT}${BOLD}A I T R I A G E${RESET}   ${GRAY}//${RESET}   ${WHITE}V I R T U A L   F O U N D R Y${RESET}\033[K\n"
    printf "  ${DIM}────────────────────────────────────────────────────────${RESET}\033[K\n"
    printf "  ${DIM}DevSecOps Security Scanner  •  Enterprise SAST Engine${RESET}\033[K\n\n"

    # Draw steps
    for (( i=0; i<$TOTAL_STEPS; i++ )); do
        if [ $i -lt $current_step ]; then
            printf "  ${GREEN}✓${RESET}   ${GRAY}${STEPS[$i]}${RESET}\033[K\n"
        elif [ $i -eq $current_step ]; then
            frame=${SPINNER[$frame_idx]}
            printf "  ${ACCENT}%s${RESET}   ${WHITE}${STEPS[$i]}${RESET}\033[K\n" "$frame"
        else
            printf "  ${DIM}·${RESET}   ${DIM}${STEPS[$i]}${RESET}\033[K\n"
        fi
    done
    printf "\033[K\n"

    # Progress bar
    pct=$(( (current_step * 100) / TOTAL_STEPS ))
    bar_width=40
    filled=$(( (pct * bar_width) / 100 ))
    empty=$(( bar_width - filled ))
    printf "  ${ACCENT}"
    for ((b=0; b<filled; b++)); do printf "█"; done
    printf "${DIM}"
    for ((b=0; b<empty; b++)); do printf "░"; done
    printf "${RESET}  ${GRAY}%d%%${RESET}\033[K\n" "$pct"

    # Advance frame
    frame_idx=$(( (frame_idx+1) % SPINNER_LEN ))
    
    # Advance step every 15 ticks (~1.5s) to simulate progress through real docker build phases
    if [ $((tick % 15)) -eq 0 ] && [ $tick -ne 0 ] && [ $current_step -lt $((TOTAL_STEPS - 1)) ]; then
        current_step=$((current_step + 1))
    fi
    
    tick=$((tick + 1))
    sleep 0.1
done

wait $PID
EXIT_CODE=$?

# Final redraw
tput rc
printf "  ${ACCENT}${BOLD}A I T R I A G E${RESET}   ${GRAY}//${RESET}   ${WHITE}V I R T U A L   F O U N D R Y${RESET}\033[K\n"
printf "  ${DIM}────────────────────────────────────────────────────────${RESET}\033[K\n"
printf "  ${DIM}DevSecOps Security Scanner  •  Enterprise SAST Engine${RESET}\033[K\n\n"

if [ $EXIT_CODE -eq 0 ]; then
    for (( i=0; i<$TOTAL_STEPS; i++ )); do
        printf "  ${GREEN}✓${RESET}   ${GRAY}${STEPS[$i]}${RESET}\033[K\n"
    done
    printf "\033[K\n"
    printf "  ${GREEN}"
    for ((b=0; b<40; b++)); do printf "█"; done
    printf "${RESET}  ${GREEN}100%%${RESET}\033[K\n"
    printf "\033[K\n"
    printf "  ${GREEN}✦${RESET}  ${WHITE}Container ready: ${ACCENT}aitriage:latest${RESET}\033[K\n"
    printf "  ${DIM}   Engines: SAST • Entropy • Secrets • NFR • Deploy Audit • Network${RESET}\033[K\n"
else
    for (( i=0; i<$TOTAL_STEPS; i++ )); do
        if [ $i -lt $current_step ]; then
            printf "  ${GREEN}✓${RESET}   ${GRAY}${STEPS[$i]}${RESET}\033[K\n"
        elif [ $i -eq $current_step ]; then
            printf "  ${RED}✖${RESET}   ${WHITE}${STEPS[$i]} [FAILED]${RESET}\033[K\n"
        else
            printf "  ${DIM}·${RESET}   ${DIM}${STEPS[$i]}${RESET}\033[K\n"
        fi
    done
    printf "\n  ${RED}✖${RESET}  ${WHITE}Build failed. Output:${RESET}\033[K\n\n"
    cat "$TMP_LOG"
fi

rm "$TMP_LOG"
tput cnorm
trap - INT TERM EXIT

