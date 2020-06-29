version: '3.3'
services:
  nats:
    image: nats-streaming:0.18-alpine
    ports:
    - 4222:4222
    - 8222:8222