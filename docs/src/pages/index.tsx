import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import styles from './index.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  // src: string;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Separation of concerns',
    Svg: require('@site/static/img/street-sign.svg').default,
    // src: "https://www.svgrepo.com/download/520485/research.svg",
    // src: "https://www.svgrepo.com/download/514345/street-sign.svg",
    // src: "/img/street-sign.svg",
    description: (
      <>
        DO was designed from the ground up to increase modularity
        by making things less connected and dependent.
      </>
    ),
  },
  {
    title: 'Application lifecycle',
    Svg: require('@site/static/img/compass.svg').default,
    // src: "https://www.svgrepo.com/download/520498/clock.svg",
    // src: "https://www.svgrepo.com/download/514347/compass.svg",
    // src: "/img/compass.svg",
    description: (
      <>
        DO provides an API to check the health of active services and to
        unload an application gracefully in reverse dependency order.
      </>
    ),
  },
  {
    title: 'Easy debugging',
    Svg: require('@site/static/img/telescope.svg').default,
    // src: "https://www.svgrepo.com/download/520483/quiz.svg",
    // src: "https://www.svgrepo.com/download/514336/telescope.svg",
    // src: "/img/telescope.svg",
    description: (
      <>
        Debugging IoC can be painful.
        DO offers an API to describe the application layout and visualize
        a service with its full dependency tree.
      </>
    ),
  },
];

function Feature({title, Svg, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        {/* <img src={src} className={styles.featureSvg} /> */}
        {/* <Svg src="" className={styles.featureSvg} role="img" /> */}
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons} style={{marginBottom: '10px'}}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/about">
            Intro
          </Link>
        </div>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/getting-started">
            Getting started - 5min ⏱️
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`⚙️ ${siteConfig.title}: ${siteConfig.tagline}`}
      description="A dependency injection toolkit based on Go 1.18+ Generics.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
