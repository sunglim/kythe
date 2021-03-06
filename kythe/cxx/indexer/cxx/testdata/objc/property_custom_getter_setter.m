// Test calling the custom getter and setter.
//
// The test assures that:
// * Regular method calls to the setter and getter methods ref/call
//   the explicit decl.
// * Setting and getting via the property ref the property decl and ref/call
//   the explicit decl.

//- @Box defines/binding BoxDecl
//- BoxDecl.node/kind record
//- BoxDecl.subkind class
//- BoxDecl.complete incomplete
@interface Box {
//- @"_testwidth" defines/binding WidthVarDecl
  int _testwidth;
}
//- @width defines/binding WidthPropDecl
@property (getter=foo,setter=bar:) int width;

//- @"foo" defines/binding FooDecl
//- FooDecl.node/kind function
//- FooDecl childof BoxDecl
-(int) foo;

//- @"bar:(int)value" defines/binding BarDecl
//- BarDecl.node/kind function
//- BarDecl childof BoxDecl
-(void) bar:(int)value;

@end

//- @Box defines/binding BoxImpl
//- BoxImpl.node/kind record
//- BoxImpl.subkind class
//- BoxImpl.complete definition
@implementation Box
@synthesize width = _testwidth;

//- @"foo " defines/binding FooImpl
//- FooImpl.node/kind function
//- FooImpl childof BoxImpl
//- @"foo " completes/uniquely FooDecl
-(int) foo {
  return self->_testwidth;
}

//- @"bar:(int)value " defines/binding BarImpl
//- BarImpl.node/kind function
//- BarImpl childof BoxImpl
//- @"bar:(int)value " completes/uniquely BarDecl
-(void) bar:(int)value {
  self->_testwidth = value;
}

@end

int main(int argc, char **argv) {
  Box *box = [[Box alloc] init];

  //- @width ref/call BarDecl
  //- @width ref WidthPropDecl
  //- @width childof Main
  //- @width.node/kind anchor
  box.width = 44;

  //- @"[box bar:55]" childof Main
  //- @"[box bar:55]".node/kind anchor
  //- @"[box bar:55]" ref/call BarDecl
  [box bar:55];

  //- @width ref/call FooDecl
  //- @width ref WidthPropDecl
  //- @width childof Main
  //- @width.node/kind anchor
  int a = box.width;

  //- @"[box foo]" ref/call FooDecl
  //- @"[box foo]" childof Main
  //- @"[box foo]".node/kind anchor
  int b = [box foo];
  return 0;
}
