# MTD data collection service

This is a very simply program to continually request departure data from <https://developer.mtd.org/documentation/v2.2/method/getdeparturesbystop/> for a set of stops, and save the results to a file in JSONL format.

## Setup

Settings are controlled via environment variables, with a `.env` file if you like. View `.env.example` to see the settings.

You will need an [API key from MTD](https://developer.cumtd.com) to use this. Also note that they have rate limits, and they explicitly say not to try to get all stops. So set the settings accordingly.

Build with `go build`.

To run as a service, copy the `mtddata.service` file to `/etc/systemd/system/mtddata.service` and edit the paths in that service file to match wherever you checked out this repository. Then run:

```
sudo systemctl enable mtddata.service
sudo systemctl start mtddata.service
sudo systemctl status mtddata.service
```
