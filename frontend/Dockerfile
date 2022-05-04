FROM node:16 AS dev

# setup okteto message
COPY bashrc /root/.bashrc

WORKDIR /usr/src/app

COPY package.json yarn.lock ./
RUN --mount=type=cache,target=/root/.yarn YARN_CACHE_FOLDER=/root/.yarn yarn install
COPY . .

RUN --mount=type=cache,target=./node_modules/.cache/webpack yarn build

FROM nginx:alpine
COPY --from=dev /usr/src/app/dist /usr/share/nginx/html
EXPOSE 80
