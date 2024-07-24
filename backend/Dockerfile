# specifies the base image for your docker image. we are using a version of the go programming language built on that distribution
FROM golang:1.16-alpine3.13

# sets working directionary inside the container to /app. all subsequent instructions will be run from this directory. if /app doesnt exist, docker will create it 
WORKDIR /app

# copies all files from the current directory (where the dockerfile is located) on your host machine into the /app directory inside the container --> essentially it includes your go application's source code in the docker image
COPY . .

#Download and install the go module dependicies specified in the go project
# -d: tells go get to only download the dependencies without installing them
# -v: enabled verbose output. provides additional information about the operations being performed by the command. This extra detail can be useful for debugging or understanding what the command is doing internally.
RUN go get -d -v ./...

#Build the go app 
#go build: compiles source code located in the current directory which is .
#-o api: specifies the output binary name to be api. after this step, the api executable will be available i nthe /app directory of your container
RUN go build -o api .

#informs docker that the container will listen on port 8000 at runtime
EXPOSE 8000

#specifies the command to run when the container starts. it tells docker to execute the api binary that was built in the previous steps, start your g o application inside the container
CMD ["./api"]


# a dockerfile is a script used by docker to automate the process of building a docker image. contains a series if instructions that docker uses to create a container image, which can then be run as a container
# for go applications, a dockerfile specifies how to set up the environment and build the go application within a docker container

#what is a docker image
#A Docker image is essentially a snapshot of a filesystem that contains everything required to run a particular application. Once an image is built, it doesn’t change. This immutability ensures that the environment remains consistent across different deployments.