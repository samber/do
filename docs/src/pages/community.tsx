import React from 'react';

import styles from './community.module.css';
import classnames from 'classnames';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';

import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

function Community() {
  const context = useDocusaurusContext();

  return (
    <Layout title="Community" description="Where to ask questions and find your soul mate">
      <header className="hero">
        <div className="container text--center">
          <h1>Community</h1>
          <div className="hero--subtitle">
            These are places where you can ask questions and find your soul mate (no promises).
          </div>
          <img className={styles.headerImg} src="/img/go-community.png" />
        </div>
      </header>
      <main>
        <div className="container">
          <div className="row margin-vert--lg">
            <div className="col text--center padding-vert--md">
              <div className="card">
                <div className="card__header">
                  <i className={classnames(styles.icon, styles.chat)}></i>
                </div>
                <div className="card__body">
                  <p>Report bugs or suggest improvements</p>
                </div>
                <div className="card__footer">
                  <Link to="https://github.com/samber/do/issues" className="button button--outline button--primary button--block">Open new issue</Link>
                </div>
              </div>
            </div>

            <div className="col text--center padding-vert--md">
              <div className="card">
                <div className="card__header">
                  <i className={classnames(styles.icon, styles.chat)}></i>
                </div>
                <div className="card__body">
                  <p>You like this project?</p>
                </div>
                <div className="card__footer">
                  <Link to="https://github.com/samber/do?tab=readme-ov-file#-contributing" className="button button--outline button--primary button--block">Start contributing!</Link>
                </div>
              </div>
            </div>

            <div className="col text--center padding-vert--md">
              <div className="card">
                <div className="card__header">
                  <i className={classnames(styles.icon, styles.twitter)}></i>
                </div>
                <div className="card__body">
                  <p>Mention &#64;samuelberthe on Twitter</p>
                </div>
                <div className="card__footer">
                  <Link to="https://twitter.com/samuelberthe" className="button button--outline button--primary button--block">Follow &#64;SamuelBerthe</Link>
                </div>
              </div>
            </div>

            <div className="col text--center padding-vert--md">
              <div className="card">
                <div className="card__header">
                  <i className={classnames(styles.icon, styles.email)}></i>
                </div>
                <div className="card__body">
                  <p>For sensitive or security related queries pop us an email</p>
                </div>
                <div className="card__footer">
                  <Link to="mailto:samuel@screeb.app" className="button button--outline button--primary button--block">samuel&#64;screeb.app</Link>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </Layout>
  );
}

export default Community;