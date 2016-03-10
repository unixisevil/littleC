/* The loop statements. */
int main()
{
  int a;
  char ch;

  /* the while */
  puts("Enter a number: ");
  a = getnum();
  while(a) {
    print(a);
    print(a*a);
    puts("");
    a = a - 1;
  }

  /* the do-while */
  puts("enter characters, 'q' to quit");
  do {
     ch = getch();
  } while(ch !='q');

  /* the for */
  for(a=0; a<10; a = a + 1) {
     print(a);
  }

  return 0;
}

