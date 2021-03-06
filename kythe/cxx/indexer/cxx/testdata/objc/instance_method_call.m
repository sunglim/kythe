// Checks that Objective-C instance methods are called via the decl.
//
// Also test that whitespace does not affect our source range for the message
// expression.

//- @Box defines/binding BoxIface
@interface Box

//- @"foo" defines/binding FooDecl
-(int) foo;
//- @"bar" defines/binding BarDecl
-(int) bar;

@end

//- @Box defines/binding BoxImpl
@implementation Box

//- @"foo " defines/binding FooDefn
//- @"foo " completes/uniquely FooDecl
-(int) foo {
  return 8;
}

//- @"bar " defines/binding BarDefn
//- @"bar " completes/uniquely BarDecl
-(int) bar {
  return 28;
}
@end

//- @main defines/binding Main
int main(int argc, char **argv) {
  Box *box = [[Box alloc] init];

  //- @"[box foo]" ref/call FooDecl
  //- @"[box foo]" childof Main
  //- @"[box foo]".node/kind anchor
  [box foo];

  //- @"[    box    bar    ]" ref/call BarDecl
  //- @"[    box    bar    ]" childof Main
  //- @"[    box    bar    ]".node/kind anchor
  [    box    bar    ]      ;

  return 0;
}

