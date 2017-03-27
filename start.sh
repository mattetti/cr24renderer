# startin arguments: http://peter.sh/experiments/chromium-command-line-switches/
# https://chromium.googlesource.com/chromium/src/+/lkgr/headless/README.md
echo "\n[start script] Starting Google Chrome at $@";
google-chrome \
--headless \
--no-sandbox \
--user-data-dir=/home/chrome/ \
--no-first-run \
--remote-debugging-port=9222 \
--disable-gpu \
--enable-logging --v=1\
--window-size=1280,1024\
$@ \
&

sleep 1;

# we need to redirect the port so we can access the remote debugging port
# from outside.
echo "\n[start script] Redirect incoming 9223 to 9222";
socat TCP-LISTEN:9223,fork TCP:127.0.0.1:9222