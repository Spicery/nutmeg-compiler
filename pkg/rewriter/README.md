# nutmeg-rewriter

## Key design ideas

- The actions operate on the current node or any descendants but not on 
  any nodes of the path. This is because the tree-walk would be much
  harder to define if the parent or sibling relationships were altered.