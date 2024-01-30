FROM golang:latest
RUN apt-get update
RUN apt-get install -y pngquant
RUN apt-get install -y -qq libtesseract-dev libleptonica-dev
ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/4.00/tessdata/
RUN apt-get install -y -qq tesseract-ocr
WORKDIR /app
COPY . .
COPY eng.traineddata /usr/share/tesseract-ocr/4.00/tessdata/
RUN go mod download
VOLUME /app
EXPOSE 8070:8070
CMD ["go", "run", "main.go"]