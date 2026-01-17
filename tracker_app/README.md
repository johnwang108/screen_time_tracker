# README

## About

This is a Wails screen-time tracking app.

TODO:
- Get total focused screen time for each day.
- For each day, get the breakdown of which apps / websites the screen time was spent on.
- Get screen time stats for past weeks, both per-day and total. Can be grouped by entire week, or just select weekday/weekend.
- For an app or category, get screen time spent on that category historically, in both per-day and total stats. Can be grouped by month, week, year, weekend, weekday, etc.


## Live Development

To run in live development mode, run `wails dev` in the project directory. In another terminal, go into the `frontend`
directory and run `npm run dev`. The frontend dev server will run on http://localhost:34115. Connect to this in your
browser and connect to your application.

## Code Checking, Linting and Formatting

To check the code, run `npm run check` in the `frontend` directory. Or run `npm run check:watch` to watch for changes.

To lint the code, run `npm run lint` in the `frontend` directory.

To format the code, run `npm run format` in the `frontend` directory.

## Building

To build a redistributable, production mode package, use `wails build`.
