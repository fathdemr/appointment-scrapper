FROM alpine:3.20

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    chromium \
    chromium-chromedriver \
    font-noto \
    font-noto-cjk \
    && rm -rf /var/cache/apk/*

ENV ZONEINFO=/usr/share/zoneinfo
# chromedp bu env var'ı okur — Alpine'de binary adı chromium-browser
ENV CHROME_PATH=/usr/bin/chromium-browser

LABEL org.opencontainers.image.title="Appointment Scrapper" \
      org.opencontainers.image.description="spor.istanbul randevu rezervasyonu botu" \
      org.opencontainers.image.authors="Fatih Demir <fath.demmr@gmail.com>"

# Non-root kullanıcı (Chrome sandbox için group id de gerekli)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY dist/linux/api api

EXPOSE 5077

ENTRYPOINT ["./api"]
