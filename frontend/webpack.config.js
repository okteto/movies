const path = require('path');
const { DefinePlugin } = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const CopyPlugin = require('copy-webpack-plugin');

const srcPath = path.join(__dirname, 'src');
const distPath = path.join(__dirname, 'dist');

module.exports = (_, argv) => {
  const mode = argv.mode ?? 'production';
  return {
    context: srcPath,
    mode,
    target: 'web',
    entry: ['./index.jsx'],
    output: {
      filename: 'app.[contenthash].js',
      path: distPath,
    },
    resolve: {
      extensions: ['.js', '.jsx', '.css'],
      modules: [
        path.resolve(path.join(__dirname, '/node_modules')),
        path.resolve(srcPath)
      ],
      alias: {
        'react-dom': '@hot-loader/react-dom'
      }
    },
    module: {
      rules: [{
        test: /\.(js|jsx)$/i,
        include: srcPath,
        use: [
          {
            loader: 'babel-loader',
            options: {
              cacheDirectory: true
            }
          }
        ]
      }, {
        test: /\.css$/i,
        include: srcPath,
        use: ['style-loader', 'css-loader']
      }, {
        test: /\.(png|jpg|gif)$/i,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 8192,
            },
          },
        ],
      }]
    },
    plugins: [
      new HtmlWebpackPlugin({
        template: './index.html',
        favicon: path.join(srcPath, 'assets/images/favicon.png')
      }),
      new CopyPlugin({
        patterns: [
          { from: './static', to: distPath },
        ],
      }),
      new DefinePlugin({
        MODE: JSON.stringify(mode),
      })
    ],
    devServer: {
      port: 80,
      host: '0.0.0.0',
      hot: true,
      allowedHosts: 'all',
      client: {
        webSocketTransport: 'ws'
      },
      webSocketServer: 'ws',
      proxy: {
        '/rent': 'http://rentals:8080/rent',
        '/rentals': 'http://rentals:8080/rentals',
        '/catalog': 'http://catalog:8080'
      },
      client: {
        webSocketURL: {
          port: 443
        },
      },
      watchFiles: {
        options: {
          usePolling: true,
        },
      },
    },
    cache: {
      type: 'filesystem',
      buildDependencies: {
        config: [__filename]
      }
    }
  };
};
