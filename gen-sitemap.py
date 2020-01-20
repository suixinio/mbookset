import datetime
base_url = "https://www.bookset.io/"
base_path = "/www/wwwroot/www.bookset.io"
#base_path = "d:/"


class Config():
    SQLALCHEMY_DATABASE_URI = 'mysql+pymysql://mbook:e3HTz5xRf4ecetL5@172.17.24.245/mbook?charset=utf8'
    #SQLALCHEMY_DATABASE_URI = 'mysql+pymysql://mbook_migrate:c6DnSf3ezwpMaTsB@182.92.105.252/mbook_migrate?charset=utf8'

from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import mapper, sessionmaker
from sqlalchemy import Table, MetaData
from sqlalchemy.orm import mapper
from operator import or_, and_

metadata = MetaData()
Base = declarative_base()
from sqlalchemy import create_engine

engine = create_engine(Config.SQLALCHEMY_DATABASE_URI)

def getModel(name, engine):
    """根据name创建并return一个新的model类
    name:数据库表名
    engine:create_engine返回的对象，指定要操作的数据库连接，from sqlalchemy import create_engine
    """
    Base.metadata.reflect(engine)
    table = Base.metadata.tables[name]
    t = type(name, (object,), dict())
    mapper(t, table)
    Base.metadata.clear()
    return t
DBSession = sessionmaker(bind=engine)
session = DBSession()

book_obj = getModel("md_books", engine)
docs_obj = getModel("md_documents", engine)
blogs_obj = getModel("md_blogs",engine)
questions_obj = getModel("so_questions",engine)

books_count = session.query(book_obj).filter_by(privately_owned=0).count()
print(books_count)
docs_count = session.query(docs_obj).count()
blogs_count = session.query(blogs_obj).count()
questions_count = session.query(questions_obj).filter(questions_obj.content_zh_cn!="",questions_obj.title_zh_cn!="").count()

books = session.query(book_obj).filter_by(privately_owned=0).all()         # 获取非私有化的书籍
#filter_books = session.query(book_obj).filter_by(privately_owned==0).all()  # 获取私有化的bookid
books_id = [i.book_id for i in books]
docs = session.query(docs_obj).filter(docs_obj.book_id.in_(books_id)).all()

blogs = session.query(blogs_obj).all()
questions = session.query(questions_obj).filter(questions_obj.content_zh_cn!="",questions_obj.title_zh_cn!="").all()

def gen_sitemapindex(count_books,count_docs,count_blogs,count_questsions):
    """
    flag
        books
        blogs
        docs
    """
    data = """<?xml version="1.0" encoding="UTF-8"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
"""

    data += gen_sitemapindex_bean(count_books,"books")
    data += gen_sitemapindex_bean(count_docs,"docs")
    data += gen_sitemapindex_bean(count_blogs,"blogs")
    data += gen_sitemapindex_bean(count_questsions,"questions")

    data += "</sitemapindex>"
    return data

def gen_sitemapindex_bean(count,flag):
    cnt = count // 9999
    cnt += 1
    data = ""
    for i in range(cnt) :
        data += "<sitemap>"
        data += "<loc>"
        data += base_url + "sitemap/"+flag+"-" + str(i) +".xml"
        data += "</loc>"
        data += "<lastmod>"
        dt = datetime.datetime.now()
        data += dt.strftime( '%Y-%m-%d' )
        data += "</lastmod>"
        data += "</sitemap>"
    return data


def gen_sitemap_book_url(book):
    """
    <url>
<loc>http://www.bookstack.cn/books/beego</loc>
<priority>0.9</priority>
<lastmod>2018-06-02 16:54:57</lastmod>
<changefreq>weekly</changefreq>
</url>
    """
    data = ""
    data += "<url><loc>"
    data += base_url + "book/"+ book.identify + "</loc>"
    data += "<priority>0.9</priority>"
    data += "<lastmod>" + book.release_time.strftime( '%Y-%m-%d' )   + "</lastmod>"
    data += "<changefreq>weekly</changefreq>"
    data += "</url>"

    return data

def gen_sitemap_books(books):
    count = 0
    index_count = 0
    fp = open(base_path+"/sitemap/books-0.xml", "w")
    fp.writelines("""<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">""")
    for i in books:
        if count > 9999:
            fp.writelines("""</urlset>""")
            count = 0
            index_count += 1
            fp = open(base_path + "/sitemap/books-"+str(index_count)+".xml" , "w")
            fp.writelines("""<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
""")
        count += 1
        data =gen_sitemap_book_url(i)
        fp.writelines(data)
    fp.writelines("""</urlset>""")

def gen_sitemap_docs(books):
    count = 0
    index_count = 0
    fp = open(base_path + "/sitemap/docs-0.xml", "w")
    fp.writelines("""<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">""")
    for book in books:
        docs = session.query(docs_obj).filter_by(book_id=book.book_id)
        for i in docs:
            if count > 9999:
                fp.writelines("""</urlset>""")
                count = 0
                index_count += 1
                fp = open(base_path + "/sitemap/docs-"+str(index_count)+".xml" , "w")
                fp.writelines("""<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
    """)
            count += 1
            data =gen_sitemap_doc_url(book.identify,i)
            fp.writelines(data)
    fp.writelines("""</urlset>""")

def gen_sitemap_doc_url(book_identify,doc):
    """
    <url>
<loc>http://www.bookstack.cn/books/beego</loc>
<priority>0.9</priority>
<lastmod>2018-06-02 16:54:57</lastmod>
<changefreq>weekly</changefreq>
</url>
    """
    data = ""
    data += "<url><loc>"
    data += base_url + "read/"+book_identify + "/"+ doc.identify + "</loc>"
    data += "<priority>0.9</priority>"
    data += "<lastmod>" + doc.modify_time.strftime( '%Y-%m-%d' )   + "</lastmod>"
    data += "<changefreq>weekly</changefreq>"
    data += "</url>"

    return data



def gen_sitemap_blogs(blogs):
    count = 0
    index_count = 0
    fp = open(base_path + "/sitemap/blogs-0.xml", "w")
    fp.writelines("""<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">""")
    for i in blogs:
        if count > 9999:
            fp.writelines("""</urlset>""")
            count = 0
            index_count += 1
            fp = open(base_path + "/sitemap/blogs-"+str(index_count)+".xml" , "w")
            fp.writelines("""<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
""")
        count += 1
        data =gen_sitemap_blog_url(i)
        fp.writelines(data)
    fp.writelines("""</urlset>""")

def gen_sitemap_blog_url(blog):
    """
    <url>
<loc>http://www.bookstack.cn/books/beego</loc>
<priority>0.9</priority>
<lastmod>2018-06-02 16:54:57</lastmod>
<changefreq>weekly</changefreq>
</url>
    """
    data = ""
    data += "<url><loc>"
    data += base_url + "blog-"+ str(blog.blog_id) + ".html"+ "</loc>"
    data += "<priority>0.9</priority>"
    data += "<lastmod>" + blog.modify_time.strftime( '%Y-%m-%d' )   + "</lastmod>"
    data += "<changefreq>weekly</changefreq>"
    data += "</url>"

    return data

def gen_sitemap_questsions(questions):
    count = 0
    index_count = 0
    fp = open(base_path + "/sitemap/questions-0.xml", "w")
    fp.writelines("""<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">""")
    for i in questions:
        if count > 9999:
            fp.writelines("""</urlset>""")
            count = 0
            index_count += 1
            fp = open(base_path + "/sitemap/questions-"+str(index_count)+".xml" , "w")
            fp.writelines("""<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
""")
        count += 1
        data =gen_sitemap_question_url(i)
        fp.writelines(data)
    fp.writelines("""</urlset>""")

def gen_sitemap_question_url(question):
    """
    <url>
<loc>http://www.bookstack.cn/books/beego</loc>
<priority>0.9</priority>
<lastmod>2018-06-02 16:54:57</lastmod>
<changefreq>weekly</changefreq>
</url>
    """
    data = ""
    data += "<url><loc>"
    data += base_url + "questions/"+ str(question.source_id) + "</loc>"
    data += "<priority>0.9</priority>"
    data += "<lastmod>" + question.created_at.strftime( '%Y-%m-%d' )   + "</lastmod>"
    data += "<changefreq>weekly</changefreq>"
    data += "</url>"

    return data



file_data = gen_sitemapindex(books_count,docs_count,blogs_count,questions_count)
with open( base_path + "/sitemap.xml", 'w') as f:
    f.write(file_data)
gen_sitemap_books(books)
gen_sitemap_docs(books)
gen_sitemap_blogs(blogs)
gen_sitemap_questsions(questions)