###### http.Handler

任何实现了接口的类型都能作为处理http请求的实例

![image-20221024230230896](D:\typora\笔记\photo\image-20221024230230896.png) 



###### day1

搭建gee框架原型

1. 定义一个结构体Engine，里面包含一个字段(姑且称为路由映射表)，为map类型，目的是存储不同的请求方法以及静态路由对应的处理方法
2. 定义构造器
3. 定义一个方法：根据请求方法和静态路由生成唯一的路径(方法:把**请求方法和静态路由拼接到一起**，这样子就是唯一的请求路径，如`GET-/hello`)，同时把路径和处理方法添加到路匹映射表
4. 定义 GET、POST方法：调用步骤3定义的方法，传入三个参数，请求方法、静态路由、处理器方法
5. 封装标准库的ListenAndServe方法
6. 实现ServeHTTP方法，通过获取请求体里的请求方法(r.Mathod)和请求的路径(r.URL.Path)，在路由映射表里面查找相应的处理器



###### day2

设计上下文环境(Context)

1. 定义结构体Context封装请求体、请求对应的响应、请求路径、请求路径以及状态码，后面三个是经常要使用的，所以直接封装在结构体中。对应的构造器方法只需要传入请求体和响应体，**请求方法和路径从请求体获取，状态码在后续方法中设置**

2. 定义Query方法获取GET请求对应的参数，POSTFORM方法获取POST请求参数

   query方法：调用请求体里的URL(表示被请求的URL)的Query方法，Query方法解析请求参数并返回匹配的values类型的变量，values定义为

   ```go
   type Values map[string][]string
   ```

   values用于查询请求参数以及构建对应的值。

   获取values变量后，调用values的Get方法，该方法获取该键对应的字符串数组里面的第一个值

    

   PostForm方法：

   调用请求体里的FormValue方法

   ![image-20221025214842875](D:\typora\笔记\photo\image-20221025214842875.png) 

   ![image-20221025214901728](D:\typora\笔记\photo\image-20221025214901728.png) 

   两者实现方式不同的原因：GET获取的是URL中的参数，因此要获取到请求体中的URL，调用URL对应的方法来获取键对应的值。而POSTFORM获取的是表单数据，因此调用的是请求体对应的查询表单的方法

3. 定义一个设置上下文结构体里状态码的方法，不仅要设置Context里面的状态码，**还要改变响应里面的状态码**(为什么改变的是响应里面的而不是请求的，因为状态码是表示响应的结果，是针对于响应状态而存在的)。调用上下文里的Writer(ResponseWriter)的`WriteHeader`方法，该方法发送HTTP回复的头域和状态码

4. 定义设置HTTP头域键值对的方法(今天主要用于设置"Content-Type"的值)

5. 定义以字符串方式生成HTTP请求响应的方法，涉及三个步骤：

   (1) 设置HTTP头域(设置"Content-Type")

   (2) 设置状态码

   (3) 调用上下文的Writer字段的Write方法，向连接中写入作为HTTP的一部分回复的数据

   **正确的调用顺序应该是Header().Set 然后WriteHeader() 最后是Write()**，可以这样子理解，要先通过Set方法设置头域键值对，然后通过WriteHeader方法把回复的头域和状态码发送出去，最后在响应体里写入部分数据，若先调用WriteHeader，即先发送了，再设置，也不会影响对应的响应体，即不会生效

6. 定义以JSON方式生成HTTP请求响应的方法，前两个步骤与第五点前两步一样，要想输出json对象，就要将json对象写入输出流，而json包里有一个将json对象写入输出流的结构体`Encoder`，需要使用NewEncoder方法创建一个Encoder结构体，传入参数为要写入json对象输出流，即上下文的响应体，最后调用该结构体的`Encode`方法，传入参数v，该方法将v的json编码写入输出流，并会写入一个换行符

7. 定义给响应体写入数据的方法，这里只需要设置状态码即可，然后调用上下文里的响应体的Write方法

8. 定义以HTML方式生成请求响应的方法，首先设置HTTP头域和状态码，然后调用调用上下文里的响应体的Write方法，往里面写入html数据(一般是string类型)

9. 把路由单独提取出来，并把`HandlerFunc`的类型从`func(http.ResponseWriter, *http.Request)`改为 `func(*Context)`，里面不仅包含了请求体和响应体，还包含了请求方法和路由路径以及状态响应码，方便快速访问这些常用的属性，并把handle方法参数改为上下文。

10. 微调框架入口。

    (1) 结构体`Engine`包含字段为值为上下文的键值对

    (2) ServeHTTP方法由直接在路由表中根据请求路由查找处理方法改为根据响应体和请求体创建一个新的上下文实例，然后调用engine结构体里的路由器结构体的handle方法，把新创建的上下文实例当作参数传进去，在handle方法里根据请求路径调用对应的处理方法



###### day3(难点)

照着打代码都能打错。。。。。

最后还是先把错误版本提交到github，然后再把正确代码一个个复制过来，通过版本控制来比较哪里不一样从而找出错误

![image-20221026011504668](D:\typora\笔记\photo\image-20221026011504668.png) 

  

今天实现了通过trie树(前缀树)是实现动态路由添加以及匹配 

1. 设计树节点的结构体，结构体包含四个四段

   1. pattern，用来表示待匹配的完整路由，例如：/hello/:name，即GET方法的第一个参数
   2. part，用来表示路由中的一部分，例如hello，:name，即完整路由根据符号 / 分隔后一段字符串，后面用来作为树节点的值
   3. children，子节点，用来存储每个节点的孩子节点，例如有两个路由，/hello/first，/hello/:name，最终在树上就会展现出，hello的children包含了两个节点，它们的值分别是first、:name
   4. isWild 布尔值，用来表示当前的节点的值是否含有 : 或 * ，这两个就是用来匹配动态路由的，若包含两者中的一个，则isWild为true

2. 编写在trie树中插入节点的方法(深度优先搜索的思路)，即通过GET方法把第一个参数添加到前缀树中，具体实现如下(调用addRoute函数)：

   1. 先把路由路径根据 / 分成几段字符串，存在变量parts里面，每一段代表一个节点的值

   2. 根据请求方法(get/post/delete等)，以请求方法为键查找路由是否存在该方法对应的根节点，若不存在则新建一个节点作为该方法对应的头节点

   3. 以查找到的头节点，调用insert方法，传入申请的完整路径、parts以及从0开始的树高(同时可以作为下标索引)

   4. insert方法首先判断树高和parts的长度是否一样，若一样则说明parts中的字段值都被添加到不同节点了，此时给调用该方法的节点(同时也是叶子节点)的pattern赋值为完整的请求路径，用来在后面查找匹配的时候表示是否匹配成功，若树高达到了要求但pattern为空说明该请求的路径在trie树中并没有对应的路径，只有树高达到要求且节点的pattern不为空，才说明在trie树中添加过和请求路径匹配的动态路径。即：`/p/:lang/doc`只有在第三层节点，即`doc`节点，`pattern`才会设置为`/p/:lang/doc`。`p`和`:lang`节点的`pattern`属性皆为空。因此，当匹配结束时，我们可以使用`n.pattern == ""`来判断路由规则是否匹配成功。例如，`/p/python`虽能成功匹配到`:lang`，但`:lang`的`pattern`值为空，因此匹配失败。

      若上述判断不为真，则不返回，根据树高的值来取parts中的值(姑且记为part)(由于树高从0开始，每次调用insert就加1，相当于就是个下标索引)，并调用当前节点的matchChild方法，传入part

      ```go
      // 第一个匹配成功的节点，用于插入
      func (n *node) matchChild(part string) *node {
      	for _, child := range n.children {
      		if child.part == part || child.isWild {
      			return child
      		}
      	}
      	return nil
      }
      ```

      遍历节点的所有孩子节点，若存在一个孩子节点的值等于part或者说孩子节点是以 / 或者 * 开头(说明是用来匹配动态路由的节点)，此时说明在该trie树中已经有相同的前缀了，返回该孩子节点，否则遍历完，返回空

      回到insert方法里面，判断刚刚的返回值，若为空，说明树种没有对应的节点，新建一个节点，节点的part字段赋值为上面的part，isWild根据part的开头是否包含 / 或者 * 来赋值真或者假，其他字段为默认值，并把该节点添加为调用该方法的孩子节点

      最后调用上述得到的孩子节点的insert方法，直到树高等于parts数组的长度，此时说明所有节点已经插入树中，并给最后一个节点的pattern赋值为该请求的完整路由路径

   5. 最后给router的字段handler里面添加上路径对应的处理方法，这里的路径是 请求方法 + “-”  + 完整路由路径

3. 编写在trie树中匹配动态路由的方法

   1. 先用变量sPath把请求的路由解析(以 / 把路由分成多段字符串)，定义一个哈希表`params`来存储动态路由(trie树种的路由)对应的请求路由里的变量情况，然后根据请求方法查找是否有对应于请求方法的根节点，若没有，则直接返回nil，nil

   2. 调用根节点的search方法，把sPath传进去，同时也把0，作为高度的参数传进去

   3. 首先判断树高和sPath的长度是否相等，相等则说明此时对于请求路由的路径匹配完毕，或者判断调用search方法的节点是否以 * 开头，若是以 * 开头，也满足情况，则执行步骤4，否则执行步骤5

   4. 判断该节点的pattern是否为空，若不为空则说明在trie树中有匹配的动态或者静态路由，对于 ：匹配来说就是sPath的长度和树高相同，对于 * 来说则不是相同，因为在路由插入的时候，在路由解析成多段字符串中有一个这样的操作

      ![image-20221027191159530](D:\typora\笔记\photo\image-20221027191159530.png) 

      则是若字段以 * 开头，则说明分段到此就行，后面的内容不会起到作用，可有可无，而在插入的方法中

      ![image-20221027191334034](D:\typora\笔记\photo\image-20221027191334034.png) 

      由于这里的parts只是包括到含有 * 字符的字符串，因此也会满足长度等于高度，即节点值以 * 开头的节点的pattern字段会被赋值为完整请求路径，该节点被视为最后一个节点

      因此sParts长度等于树高且节点的pattern不为空或者节点值以 * 开头会返回调用该方法的节点为结束，而若pattern为空说明trie树其实并没有注册相同长度的路由，因此返回空

      执行完步骤4后执行步骤6

   5. 以height为下标索引取出sParts中对应的值，并把此作为参数调用该节点的matchChildren方法

       ```go
       func (n *node) matchChildren(part string) []*node {
       	nodes := make([]*node, 0)
       	for _, child := range n.children {
       		if child.part == part || child.isWild {
       			nodes = append(nodes, child)
       		}
       	}
       	return nodes
       }
       ```

      遍历调用该方法的孩子节点，判断该孩子节点的值是否和传入进来的字段值一样，若一样则说明找到相同前缀，则把其加入到`nodes数组`中(用来保存成功匹配的节点)，或者该节点的isWild字段是否为true，因为若为true，则说明该节点是动态匹配请求路由的节点，也相当于是有相同前缀，因此也加入到nodes数组中。最后返回nodes数组，回到search中

      search获取到nodes数组中，遍历该数组，并且调用里面的值的search方法，依旧传入sParts数组，传入的高度加1。即相当于进入下一层，不断深搜，不断调用孩子节点的search方法，直到进入符合进入步骤4的条件。上一层的search接收到下一层的search传来的节点后，先判断是否为空，若为空，则说明该trie路径并不匹配请求的路径，继续遍历nodes数组，若不为空则继续返回节点到上一层search，直到最初的search

   6. search返回的是注册到trie树中的完整路径对应的那个节点，例如：/hello/:name 。步骤2调用完search后会获取到一个节点，若该节点为空，则直接返回nil，nil，说明匹配失败。否则，获取到节点的字段pattern则是完整的注册路径，然后对该注册路径，以 / 为分隔符进行解析，分层多段字符串，并且对解析结果进行遍历，由于注册的路由解析后和请求的路由解析后长度一样，所以遍历解析的注册路由相当于遍历解析的请求路由。若其中一段字符串以 ：开头，则说明该字符串是用来匹配动态路由的，此时去掉 ：字符，并且以这个注册路由中的字符串为键，请求路由中对应的字符串为值，添加到步骤1定义的哈希表`params`中，然后继续遍历解析后的注册路由。若字符串以 * 开头，依旧是以去掉 * 字符的该字符串为键，以从遍历位置开始(包括)，把剩余的解析后的请求路由中的字符串全部用 / 连接起来作为值添加到`params`中，然后退出循环。循环结束后把search获取到的节点以及params返回到handle中(一开始是通过handle方法调用匹配动态路由的方法)

   7. 判断步骤6中获取的节点是否为空，若不为空则把步骤6中接收到的params赋值给Context新加的params字段，**该字段就是提供对路由参数的访问**，并且把请求的方法和请求的路径通过 - 连接到一起，并以此为键去router中寻找对应处理方法，并把传入的上下文传入该处理方法中以此来调用该处理方法。若节点为空则向浏览器中输出一条信息表示路径未匹配到。

4. 在router结构体中添加一个字段`roots`来存储每种请求方式的Trie 树根节点

5. 对Context对象增加一个`字段Params`，来提供对路由参数的访问，

   增加一个`方法Param`，来提供通过动态路由中的参数来获取匹配到的请求路由中的值(和gin一样)