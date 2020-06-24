package collect

// func makeJsonChecker2(code string, ctx *Context) (Checker, error) {
//   var exp jsonExpression2
//   if e := json.Unmarshal([]byte(code), &exp); nil != e {
//     return nil, errors.New("unmarshal expression failed, " + e.Error())
//   }
//   if "" == exp.Operator {
//     return nil, errors.New("'operator' of expression is required.")
//   }
//   if "" == exp.Value {
//     return nil, errors.New("'value' of expression is required.")
//   }

//   if "" != exp.Function {
//     exp.Function = strings.ToLower(exp.Function)
//   }

//   is_all_match := false
//   if "all" == exp.Function || "all()" == exp.Function {
//     is_all_match = true
//     exp.Function = ""
//   } else if "any" == exp.Function || "any()" == exp.Function {
//     is_all_match = false
//     exp.Function = ""
//   }

//   isCount := exp.Function == "count"

//   resolver, e := createResolver(ctx, exp.Filters, exp.Attribute, exp.Function)
//   if nil != e {
//     return nil, e
//   }

//   matcher, e := sampling.MakeMatcher(resolver.ResultType().Type, exp.Operator, exp.Value)
//   if nil != e {
//     return nil, e
//   }
//   return &CollectChecker{
//     match: func(value interface{}, ctx *CheckContext) (bool, error) {
//       return matcher.Call(true, value)
//     },
//     isCount:      isCount,
//     is_all_match: is_all_match,
//     resolver:     resolver}, nil
// }

// func createResolver(ctx *Context, filters []sampling.ExprSpec, attribute, function string) (sampling.Resolver, error) {
//   if nil == ctx {
//     return nil, errors.New("'ctx' is required.")
//   }
//   obj := ctx.sampling_client
//   if nil == obj {
//     return nil, ErrSamplingClientIsRequired
//   }
//   if channelClient, ok := obj.(sampling.ChannelClient); ok {
//     return channelClient.CreateResolver(filters, "", attribute, function)
//   } else if client, ok := obj.(sampling.Client); ok {
//     return client.CreateResolver(filters, "", attribute, function)
//   }
//   return nil, errors.New("'sampling_client' isn't sampling.ChannelClient or sampling.Client.")
// }
