module.exports = {
    "env": {
        "browser": true,
        "es6": true
    },
    "parserOptions": {
        "ecmaVersion": 2017,
        "ecmaFeatures": {
            "experimentalObjectRestSpread": true,
            "jsx": true
        },
        "sourceType": "module"
    },
    "ignorePatterns": [
        "*.min.js"
    ],
    "extends": "eslint:recommended",
    "rules": {
        "no-unused-vars": "off",
        "semi": ["error", "always"],
        "comma-dangle": ["error", "always-multiline"]
    }
};
