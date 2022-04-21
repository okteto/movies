FROM node:16

WORKDIR /src

COPY package.json yarn.lock ./
RUN --mount=type=cache,target=/root/.yarn YARN_CACHE_FOLDER=/root/.yarn yarn install

COPY . .
CMD ["yarn", "start"]
