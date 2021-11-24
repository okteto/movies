FROM node:14

# setup okteto message
COPY bashrc /root/.bashrc

WORKDIR /usr/src/app

COPY package.json yarn.lock ./
RUN --mount=type=cache,target=/root/.yarn YARN_CACHE_FOLDER=/root/.yarn yarn install

COPY . .
CMD ["yarn", "start"]
