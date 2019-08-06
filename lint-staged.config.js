module.exports = {
  '*.(js|jsx|ts|tsx)': ['eslint'],
  '*.+(js|jsx|json|yml|yaml|css|less|scss|ts|tsx|md|graphql|mdx)': [
    'prettier --write',
    'git add'
  ]
}
