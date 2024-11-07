# plot_retries.gp
set terminal pngcairo size 800,600
set output 'retries_plot.png'

set title "Retry Durations Across Strategies"
set xlabel "Attempt Number"
set ylabel "Duration (seconds)"

set key outside
set grid

plot "retries.txt" using 1:2 with linespoints title "Linear Retry" lt rgb "blue", \
     "retries.txt" using 1:3 with linespoints title "Exponential Backoff" lt rgb "green", \
     "retries.txt" using 1:4 with linespoints title "Capped Duration" lt rgb "red"
