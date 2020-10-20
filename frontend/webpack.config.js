const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const appPath = path.join(__dirname, '/src');

module.exports = {
  context: appPath,
  mode: 'development',
  devtool: 'source-map',
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
      exclude: /node_modules/
    }, {
      test: /\.css$/,
      use: ['style-loader', 'css-loader']
    }, {
      test: /\.(png|jpg|svg)$/,
      use: [
        {
          loader: 'url-loader',
          options: {
            limit: 100000,
          },
        }
      ]
    }]
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: './index.html'
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
