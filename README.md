# Go Image Filter

A simple image filtering program.

See `./img` for samples

Run `go run -h` for instructions

## Row filter
Replace each row in the image with the min, avg or max color value in the original row.

![Max value by row](img/P1490270-row-max-high-255-low-0.png)

That's boring, let's add a shadow mask.

![That's better!](img/P1490270-row-max-high-255-low-10.png)

Or sort the colors in each row.

![Pixel sort](img/P1490270-row-sort-high-255-low-16.png)

## Checkerbox filter
![Checkboxer filter](img/P1490270-checker-77-min-high-255-low-0.png)

With shadow and highlight massks

![Checkboxer filter](img/P1490270-checker-77-max-high-184-low-84.png)

# More samples!

![Checkboxer sky!](img/P1490239-checker-82-max-high-212-low-33.png)

![On the beach](img/P1490201-row-max-high-255-low-72.png)