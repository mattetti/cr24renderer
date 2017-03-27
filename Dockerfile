
FROM ubuntu

WORKDIR /tmp

RUN useradd -m chrome
RUN apt-get update
RUN apt-get install -y wget
RUN apt-get install -y unzip

RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN wget https://chromedriver.storage.googleapis.com/2.28/chromedriver_linux64.zip

RUN apt-get install -y gconf-service
RUN apt-get install -y libasound2 libatk1.0-0 libcairo2 libcups2 libfontconfig1 libfreetype6 libgdk-pixbuf2.0-0 libgtk2.0-0 libnspr4 libnss3
RUN apt-get install -y libpango1.0-0 libxss1 libxtst6 libappindicator1 libcurl3 xdg-utils
RUN apt-get install -y fonts-liberation

# use socat to redirect the 9222 port which isn't available outside of localhost
RUN apt-get install -y socat

RUN dpkg -i /tmp/google-chrome-stable_current_amd64.deb
RUN unzip chromedriver_linux64.zip
RUN mv chromedriver /usr/local/bin/chromedriver
RUN rm -rf /tmp/*
RUN apt-get purge
RUN apt-get clean
RUN apt-get autoremove -y
USER chrome
WORKDIR /home/chrome
COPY start.sh /home/chrome/

ENTRYPOINT ["sh", "/home/chrome/start.sh"]
# chrome debugger
EXPOSE 9223
# chromedriver (not used at the moment)
EXPOSE 9515