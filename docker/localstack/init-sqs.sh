#!/bin/sh

awslocal sqs create-queue \
  --queue-name chat-events