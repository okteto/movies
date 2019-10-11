FROM node:10 as builder

WORKDIR /src

COPY package.json yarn.lock ./
RUN yarn install
COPY . .

RUN yarn build

FROM nginx:alpine 
COPY --from=builder /src/dist /usr/share/nginx/html
EXPOSE 80


