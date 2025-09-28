# Examples of the nutmeg-parser in use

## Simple example

A trivial definition in Nutmeg:
```
### A simple definition 
def incr(x):
    x + 1
end
```

We can parse that with a simple command-line and see the results in XML:

`❯ cat incr.nutmeg | nutmeg-tokenizer | nutmeg-parser -f xml`
```xml
<unit span="1 1 3 4">
  <form syntax="surround" span="1 1 3 4">
    <part keyword="def" span="1 1 1 12">
      <apply kind="parentheses" span="1 5 1 12">
        <identifier name="incr" span="1 5 1 9" />
        <delimited kind="parentheses" span="1 9 1 12">
          <identifier name="x" span="1 10 1 11" />
        </delimited>
      </apply>
    </part>
    <part keyword="=&gt;&gt;" span="1 12 2 10">
      <operator name="+" syntax="infix" span="2 5 2 10">
        <identifier name="x" span="2 5 2 6" />
        <number base="10" exponent="0" fraction="" mantissa="1" span="2 9 2 10" />
      </operator>
    </part>
  </form>
</unit>
```
