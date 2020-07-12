name: CI
on: "push"

jobs:
  test:
    runs-on: ubuntu-latest
    container:
      image: golang:1.14
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Cache Go Dependencies
        uses: actions/cache@v2
        id: go-dep-cache
        with:
          path: /home/worker/go/pkg
          key: v1-${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            v1-${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
            v1-${{ runner.os }}-go-
      - name: Download Dependencies
        run: make dep
      - name: Run Tests
        run: make test

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Set up Docker Buildx
        uses: crazy-max/ghaction-docker-buildx@v3
      - name: Cache Docker layers
        uses: actions/cache@v2
        id: buildx-cache
        with:
          path: /tmp/.buildx-cache
          key: v1-${{ runner.os }}-buildx-${{ hashFiles('**/Dockerfile') }}
          restore-keys: |
            v1-${{ runner.os }}-buildx-${{ hashFiles('**/Dockerfile') }}
            v1-${{ runner.os }}-buildx-
      - name: Build and Push Docker Container
        env:
          IMAGE_PUSH_SECRET: ${{ secrets.DOCKER_IMAGE_PUSH }}
        run: .ci/docker-builder.sh
