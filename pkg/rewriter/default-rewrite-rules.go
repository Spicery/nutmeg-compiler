package rewriter

const DefaultRewriteRules = `
name: Default Rewrite
passes:

  - name: Pass 1
    singlePass: true
    downwards:

      - name: (x.f)(y) -> f(x,y)
        match:
          self:
            name: apply
            count: 2
          child:
            name: operator
            key: name
            value: "."
            count: 2
            siblingPosition: 0
          nextChild:
            name: arguments
            key: kind
            value: parentheses
        action:
          sequence:
            - replaceName: 
                with: apply
            - inlineChild: true
            - permuteChildren: [0, 1]
            - newNodeChild:
                name: arguments
                key: kind
                value: parentheses
                offset: 1
                length: 2

      - name: Rename negation as seq and toggle sign of number
        match:
          self:
            name: operator
            key: name
            value: "-"
            count: 1
          child:
            name: number
        action:
          sequence:
          - replaceName:
              with: seq
          - childAction:
              rotateOption:
                key: sign
                values: ["+", "-"]

      - name: Operator to Syscall
        match:
          self:
            name: operator
            key: name
            value.regexp: "[-+*/<>]|==|<=|>=|\\.\\.[<=]"
            count: 2
        action:
          replaceName:
            with: syscall

      - name: Infix colon
        match:
          self:
            name: operator
            key: name
            value: ":"
            count: 2
        action:
          sequence:
          - replaceName:
              with: syscall
          - replaceValue:
              key: name
              with: "=>"

      - name: Change ':=' to bind (POP)
        match:
          self:
            name: operator
            key: name
            value: ":="
            count: 2
        action:
          replaceName:
            with: bind

      - name: Change '<-' to assign (POP)
        match:
          self:
            name: operator
            key: name
            value: "<-"
            count: 2
        action:
          replaceName:
            with: assign

      - name: Change '<--' to update (UCALL)
        match:
          self:
            name: operator
            key: name
            value: "<--"
            count: 2
        action:
          replaceName:
            with: update

      - name: Fuse let parts
        match:
          self:
            name: form
            count: 2
          child:
            name: part
            key: keyword
            value: let
            siblingPosition: 0
        action:
          mergeChildWithNext: false

      - name: Switch
        match:
          self:
            name: form
          child:
            name: part
            key: keyword
            value: switch
            siblingPosition: 0
        action:
          replaceName:
            with: switch

      - name: remove endcase
        match:
          self:
            name: switch
          child:
            name: part
            key: keyword
            value: endcase
        action:
          removeChild: true

      - name: mark else part as default
        match:
          parent:
            name: switch
          self:
            name: part
            key: keyword
            value: else
        action:
          replaceName:
            with: default

      - name: mark case and then parts
        match:
          parent:
            name: switch
          self:
            name: part
            key: keyword
            value.regexp: case|then
        action:
          replaceName:
            src: self
            from: value

  - name: Pass 2, Conditional, handle ifnot/elseifnot
    singlePass: true
    downwards:

      - name: part/if
        match:
          self:
            name: part
            key: keyword
            value: if
        action:
          replaceValue:
            key: iftype
            with: if

      - name: part/elseif
        match:
          self:
            name: part
            key: keyword
            value: elseif
        action:
          replaceValue:
            key: iftype
            with: elseif

      - name: part/ifnot
        match:
          self:
            name: part
            key: keyword
            value: ifnot
        action:
          replaceValue:
            key: iftype
            with: if

      - name: part/elseifnot
        match:
          self:
            name: part
            key: keyword
            value: elseifnot
        action:
          replaceValue:
            key: iftype
            with: elseif

    upwards:
      - name: form->if
        match:
          self:
            name: form
          child:
            name: part
            key: iftype
            value: if
            siblingPosition: 0
        action:
          replaceName:
            with: if

      - name: add-else
        match:
          self:
            name: if
          child:
            name: part
            key: keyword
            value: else
            cmp: false
            siblingPosition: -1
        action:
          newNodeChild:
            name: part
            key: keyword
            value: else
            offset: 1
            length: 0

      - name: normalise-elseif
        match:
          self:
            name: if
          child:
            name: part
            key: iftype
            value: elseif
            siblingPosition: -3
        action:
          newNodeChild:
            name: if
            length: 3
        repeatOnSuccess: true

  - name: Pass 3, Conditional, introduce ifnot
    singlePass: true
    downwards:
      - name: if-not
        match:
          self:
            name: if
          child:
            name: part
            key: keyword
            value: ifnot
            siblingPosition: 0
        action:
          replaceName:
            with: ifnot

      - name: elseif-not
        match:
          self:
            name: if
          child:
            name: part
            key: keyword
            value: elseifnot
            siblingPosition: 0
        action:
          replaceName:
            with: ifnot

  - name: Pass 4, convert forms to seq
    downwards:

      - name: Rename form using keyword
        match:
          self:
            name: form
          child:
            name: part
            key: keyword
            siblingPosition: 0
        action:
          replaceName:
            src: child
            from: value

      - name: Normalise fn, delimited->arguments
        match:
          self:
            name: part
            key: keyword
            value: fn
          child:
            name: delimited
            key: kind
            value: parentheses
            siblingPosition: 0
        action:
          childAction:
            replaceName:
              with: arguments

      - name: Normalise fn, id->arguments
        match:
          self:
            name: part
            key: keyword
            value: fn
            count: 1
          child:
            name: id
            key: name
        action:
          newNodeChild:
            name: arguments
            offset: 0
            length: 1

      - name: Rename parts to seq
        match:
          self:
            name: part 
        action:
          replaceName:
            with: seq

      - name: Parentheses to seq
        match:
          self:
            name: delimited
            key: kind
            value: parentheses
        action:
          replaceName:
            with: seq

    upwards:
      - name: Inline nested sequences
        match:
          self:
            name: seq
          child:
            name: seq
        action:
          inlineChild: true
        repeatOnSuccess: true

      - name: seq-1
        match:
          self:
            name: seq
            count: 1
        action:
          replaceByChild: 0
        onSuccess: Inline nested sequences

      - name: def->bind
        match:
          parent:
            name: def
            count: 2
          self:
            name: apply
            count: 2
          child:
            name: id
        action:
          childAction:
            replaceValue:
              key: protected
              with: "true"

      - name: def->bind
        match:
          self:
            name: def
            count: 2
          child:
            name: apply
            count: 2
        action:
          sequence:
            - replaceName:
                with: bind
            - inlineChild: true
            - newNodeChild:
                name: fn
                offset: 1
                length: 2

  - name: Pass 5, qualifiers
    singlePass: true

    downwards:
      - name: Rewrite qualifiers
        match:
          self:
            name: var
          child:
            name: id
        action:
          sequence:
            - childAction:
                replaceValue:
                  key: var
                  with: "true"
            - childAction:
                replaceValue:
                  key: const
                  with: "false"
            - replaceByChild: 0
        breakOnSuccess: true

      - name: Rewrite qualifiers
        match:
          self:
            name: val
          child:
            name: id
        action:
          sequence:
            - childAction:
                replaceValue:
                  key: var
                  with: "false"
            - childAction:
                replaceValue:
                  key: const
                  with: "false"
            - replaceByChild: 0
        breakOnSuccess: true

      - name: Rewrite qualifiers
        match:
          self:
            name: const
          child:
            name: id
        action:
          sequence:
            - childAction:
                replaceValue:
                  key: var
                  with: "false"
            - childAction:
                replaceValue:
                  key: const
                  with: "true"
            - replaceByChild: 0
        breakOnSuccess: true

      - name: Validate qualifiers
        match:
          self:
            name: val
        action:
          fail: "Qualifier was not followed by an identifier"

    upwards:
      - name: Remove syntax=VALUE
        match:
          self:
            key: syntax
        action:
          removeOption: 
            key: syntax

      - name: Remove bind options
        match:
          self:
            name: bind
        action:
          clearOptions: true
`
