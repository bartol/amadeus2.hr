import React from 'react';
import { graphql } from 'gatsby';
import { Layout } from '../components/layout';
import { Card } from '../components/card';

const Index = ({ pageContext, data }) => {
    const { language } = pageContext;
    const { items } = data.amadeus;

    return (
        <Layout language={language}>
            {/* TODO */}
            <ul>
                {items.map(item => {
                    return (
                        <Card
                            name={item.name}
                            price={item.price}
                            discountedPrice={item.discountedPrice}
                            image={item.images[0]}
                            type={item.type}
                            quantity={item.quantity}
                            availability={item.availability}
                            slug={item.slug}
                            id={item.id}
                            hidden={item.hidden}
                            key={item.id}
                        />
                    );
                })}
            </ul>
        </Layout>
    );
};

export default Index;

export const indexQuery = graphql`
    query indexQuery {
        amadeus {
            items {
                name {
                    hr
                    en
                }
                price
                discountedPrice
                slug
                id
                type {
                    hr
                    en
                }
                brand
                quantity
                availability {
                    hr
                    en
                }
                images {
                    index
                    src
                    optimized {
                        childImageSharp {
                            # cards
                            fluid(maxWidth: 240, maxHeight: 180) {
                                ...GatsbyImageSharpFluid_withWebp
                            }
                            # cart
                            fixed(width: 120, height: 90) {
                                ...GatsbyImageSharpFixed_withWebp
                            }
                        }
                    }
                }
            }
        }
    }
`;
