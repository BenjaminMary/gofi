# Docker Desktop

# go inside the repo with the Dockerfile : cd C:\git\go\gosheets
# build image : docker build --tag imageName1 .
# build image : docker build --tag benjaminmary/gosheets:port8082 .
# The --tag flag tags your image with a name. And the . lets Docker know where it can find the Dockerfile, here on the same folder. Format = repo/name:tag

# Once the build is complete, an image will appear in the Images tab. Select the image name to see its details. 

# run the image in a container : Select Run inside the image details. In the Optional settings remember to specify a port number (something like 8080+).
# low MB image : https://medium.com/@pavelfokin/how-to-build-a-minimal-golang-docker-image-b4a1e51b03c8

# go images : https://hub.docker.com/_/golang
FROM golang:1.21.1-alpine

WORKDIR /usr/src/app

# RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the source files into the image.
COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server

EXPOSE 8082

# Run the application.
# CMD go run .
CMD ["/server"]