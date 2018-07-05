FROM phusion/baseimage
WORKDIR /app
COPY . /app/
ENTRYPOINT ["/app/pdf2dash"]
EXPOSE 3000