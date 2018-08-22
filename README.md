# alexa-smart-home

## Description

Go libraries that can be used to build a personal or public Alexa smart home skill (https://developer.amazon.com/docs/smarthome/steps-to-build-a-smart-home-skill.html).

The initial skill entrypoint must run as an AWS Lambda function but asynchronous responses are supported so you can offload the processing elsewhere.

## Usage

Check out the example package for samples of building the initial Lambda function and offloading handling via SQS. Only a subset of devices are currently implemented but it should be possible to add your own.