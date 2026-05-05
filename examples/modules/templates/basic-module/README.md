# Basic PE Module Template

Use this directory as a small starting point for a prompt module.

```sh
cp -R examples/modules/templates/basic-module my-prompts
cd my-prompts
pe mod tidy --write
pe mod vet prompts/review.pe
```

The template uses executable text so the prompt can declare inputs and policy
without losing plain-text readability.
