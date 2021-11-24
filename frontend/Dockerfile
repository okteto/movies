FROM node:14 as dev

WORKDIR /usr/src/app

COPY package.json yarn.lock ./
RUN yarn install

COPY . .
RUN --mount=type=cache,target=./node_modules/.cache/webpack yarn build

FROM nginx:alpine
COPY --from=dev /usr/src/app/dist /usr/share/nginx/html
EXPOSE 80
