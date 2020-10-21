module.exports = {
    "env": {
        "browser": true,
        "es6": true
    },
    "parserOptions": {
        "ecmaFeatures": {
            "experimentalObjectRestSpread": true,
            "jsx": true
        },
        "sourceType": "module"
    },
    "extends": "eslint:recommended",
    "rules": {
        "no-unused-vars": "off",
        "semi": ["error", "always"],
        "comma-dangle": ["error", "always-multiline"]
    }
};
