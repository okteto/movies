FROM node:20

# Enable Corepack and prepare Yarn 4.9.4
RUN corepack enable && corepack prepare yarn@4.9.4 --activate

# Copy dependencies
WORKDIR /src
COPY package.json yarn.lock ./

# Install using Yarn 4 with node_modules (disable PnP)
RUN --mount=type=cache,target=/root/.yarn \
    YARN_CACHE_FOLDER=/root/.yarn \
    yarn config set nodeLinker node-modules && \
    yarn install --immutable

COPY . .
CMD ["yarn", "start"]
