const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const appPath = path.join(__dirname, '/src');

module.exports = {
  context: appPath,
  target: 'web',
  mode: 'development',
  devtool: 'eval-cheap-source-map',
  entry: ['./index.jsx'],
  output: {
    filename: 'app.[contenthash].js',
    path: path.resolve(__dirname, '/dist'),
  },
  resolve: {
    extensions: ['.js', '.jsx', '.css'],
    modules: [
      path.resolve(path.join(__dirname, '/node_modules')),
      path.resolve(appPath)
    ]
  },
  module: {
    rules: [{
      test: /\.(js|jsx)$/,
      use: 'babel-loader',
      include: path.resolve(__dirname, 'src')
    }, {
      test: /\.css$/,
      use: ['style-loader', 'css-loader'],
      include: path.resolve(__dirname, 'src')
    }, {
      test: /\.(png|jpg|svg)$/,
      use: [
        {
          loader: 'url-loader',
          options: {
            limit: 100000,
          },
        }
      ],
      include: path.resolve(__dirname, 'src')
    }]
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: './index.html',
      favicon: path.resolve(__dirname, 'src/assets/images/favicon.png')
    })
  ],
  devServer: {
    port: 80,
    host: '0.0.0.0',
    hot: true,
    sockPort: 443,
    disableHostCheck: true,
    watchOptions: {
      poll: true
    },
    proxy: {
      '/api': 'http://movies-api:8080'
    }
  }
};
